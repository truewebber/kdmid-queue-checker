package daemon

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"kdmid-queue-checker/domain/captcha"
	"kdmid-queue-checker/domain/crawl"
	"kdmid-queue-checker/domain/log"
	"kdmid-queue-checker/domain/notification"
	"kdmid-queue-checker/domain/page"
)

type CheckSlot struct {
	dispatcher       page.Dispatcher
	solver           captcha.Solver
	crawlStorage     crawl.Storage
	recipientStorage notification.Storage
	notifier         notification.Notifier
	logger           log.Logger
}

func NewCheckSlot(
	dispatcher page.Dispatcher,
	solver captcha.Solver,
	crawlStorage crawl.Storage,
	recipientStorage notification.Storage,
	notifier notification.Notifier,
	logger log.Logger,
) *CheckSlot {
	return &CheckSlot{
		dispatcher:       dispatcher,
		solver:           solver,
		crawlStorage:     crawlStorage,
		recipientStorage: recipientStorage,
		notifier:         notifier,
		logger:           logger,
	}
}

func (c *CheckSlot) Handle(ctx context.Context) error {
	const maxAmountOfTriggersADay = 20

	checkFrom := "05:00"
	checkTil := "20:00"

	cronRules, err := c.getCronRules(checkFrom, checkTil, maxAmountOfTriggersADay)
	if err != nil {
		return fmt.Errorf("get cron rules: %w", err)
	}

	cr := cron.New(cron.WithSeconds())

	for _, cronRule := range cronRules {
		if _, err := cr.AddFunc(cronRule, func() {
			c.runAllRecipients(ctx)
		}); err != nil {
			return fmt.Errorf("cron add func: %w", err)
		}

		c.logger.Info("rule registered", "_", cronRule)
	}

	cr.Start()

	select {
	case <-ctx.Done():
		exitCtx := cr.Stop()
		<-exitCtx.Done()

		return nil
	}
}

func (c *CheckSlot) parseTime(timeStr string) (int, int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	if hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour: %d", hour)
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	if minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute: %d", minute)
	}

	return hour, minute, nil
}

func (c *CheckSlot) getCronRules(startCheckFrom, startCheckTil string, amountOfTriggersADay int) ([]string, error) {
	if amountOfTriggersADay <= 0 {
		return nil, fmt.Errorf("amountOfTriggersADay must be greater than 0")
	}

	startHour, startMinute, err := c.parseTime(startCheckFrom)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}

	endHour, endMinute, err := c.parseTime(startCheckTil)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	startTime := time.Date(0, time.January, 1, startHour, startMinute, 0, 0, time.UTC)
	endTime := time.Date(0, time.January, 1, endHour, endMinute, 0, 0, time.UTC)

	if endTime.Before(startTime) {
		endTime = endTime.Add(24 * time.Hour)
	}

	totalMinutes := int(endTime.Sub(startTime).Minutes())
	if totalMinutes == 0 {
		totalMinutes = 24 * 60
	}

	intervalMinutes := 1

	if amountOfTriggersADay > 1 {
		intervalMinutes = totalMinutes / (amountOfTriggersADay - 1)
		if intervalMinutes < 1 {
			return nil, fmt.Errorf("intervalMinutes must be greater than or equal to 1")
		}
	}

	cronRules := make([]string, 0, amountOfTriggersADay)

	for i := 0; i < amountOfTriggersADay; i++ {
		triggerTime := startTime.Add(time.Duration(i*intervalMinutes) * time.Minute)
		cronRule := fmt.Sprintf("%d %d %d * * *", triggerTime.Second(), triggerTime.Minute(), triggerTime.Hour())
		cronRules = append(cronRules, cronRule)
	}

	return cronRules, nil
}

func (c *CheckSlot) runAllRecipients(ctx context.Context) {
	recipients, err := c.recipientStorage.List(ctx)
	if err != nil {
		c.logger.Error("list recipients failed", "err", err)
	}

	for _, recipient := range recipients {
		if err := c.runSingleCheck(ctx, &recipient); err != nil {
			c.logger.Error("check slot failed", "recipient", recipient, "err", err)
		}
	}
}

func (c *CheckSlot) runSingleCheck(
	ctx context.Context,
	recipient *notification.Recipient,
) error {
	c.logger.Info("start run single check")

	crawlResult, crawlErr := c.crawl(recipient.ID, recipient.CD)
	if crawlErr != nil {
		return fmt.Errorf("crawl failed: %w", crawlErr)
	}

	if saveErr := c.crawlStorage.Save(ctx, recipient.TelegramID, crawlResult); saveErr != nil {
		crawlResult.Err = fmt.Errorf("%w, save crawl result: %w", crawlResult.Err, saveErr)

		c.logger.Error("save crawl result", "recipient", recipient, "err", saveErr)
	}

	if crawlResult.SomethingInteresting || crawlResult.Err != nil {
		if notifyErr := c.notify(ctx, crawlResult, recipient); notifyErr != nil {
			return fmt.Errorf("notify failed: %w", notifyErr)
		}
	}

	c.logger.Info("run single check finished", "something_interesting", crawlResult.SomethingInteresting)

	return nil
}

func (c *CheckSlot) notify(
	ctx context.Context,
	result *crawl.Result,
	recipient *notification.Recipient,
) error {
	n := c.buildNotification(result)

	if err := c.notifier.Notify(ctx, n, recipient); err != nil {
		c.logger.Error("notify failed", "err", err)
	}

	return nil
}

func (c *CheckSlot) buildNotification(result *crawl.Result) *notification.Notification {
	return &notification.Notification{
		Images:               c.buildNotificationImages(result),
		CrawledAt:            result.RanAt,
		Error:                result.Err,
		SomethingInteresting: result.SomethingInteresting,
	}
}

func (c *CheckSlot) buildNotificationImages(result *crawl.Result) []notification.PNG {
	images := make([]notification.PNG, 0)

	for _, stat := range []page.Stat{result.One, result.Two, result.Three} {
		if len(stat.Screenshot) != 0 {
			images = append(images, notification.PNG(stat.Screenshot))
		}

		if stat.Captcha.Presented {
			images = append(images, notification.PNG(stat.Captcha.Image))
		}
	}

	return images
}

func (c *CheckSlot) crawl(applicationID, applicationCD string) (*crawl.Result, error) {
	i := 0

	for {
		crawlResult, err := c.crawlWithRetry(applicationID, applicationCD, i)
		if errors.Is(err, errRetryCrawl) {
			i++

			c.logger.Info("retry crawl", "idx", i, "err", err)

			continue
		}

		if err != nil {
			return nil, fmt.Errorf("crawl failed, retryIdx - %d: %w", i, err)
		}

		return crawlResult, nil
	}
}

var errRetryCrawl = fmt.Errorf("retry crawl")

const maxRetryOnCaptchaNotSolved = 3

func (c *CheckSlot) crawlWithRetry(applicationID, applicationCD string, retryIdx int) (*crawl.Result, error) {
	navigator, err := c.dispatcher.NewNavigator(applicationID, applicationCD)
	if err != nil {
		return nil, fmt.Errorf("new navigator: %w", err)
	}

	defer c.logger.CloseWithLog(navigator)

	crawlResult := &crawl.Result{
		RanAt: time.Now(),
	}

	crawlResult.One, err = navigator.OpenPageToAuthorize()
	if err != nil {
		crawlResult.Err = fmt.Errorf("open page to authorize: %w", err)

		return crawlResult, nil
	}

	code, err := c.solver.Solve(crawlResult.One.Captcha.Image)
	if err != nil {
		crawlResult.Err = fmt.Errorf("solve captcha: %w", err)

		return crawlResult, nil
	}

	crawlResult.Two, err = navigator.SubmitAuthorization(code)
	if errors.Is(err, page.ErrCaptchaNotSolved) && retryIdx < maxRetryOnCaptchaNotSolved {
		return nil, fmt.Errorf("%w: %w", err, errRetryCrawl)
	}

	if err != nil {
		crawlResult.Err = fmt.Errorf("submit authorization: %w", err)

		return crawlResult, nil
	}

	crawlResult.Three, err = navigator.OpenSlotBookingPage()
	if err != nil {
		crawlResult.Err = fmt.Errorf("open slot booking page: %w", err)

		return crawlResult, nil
	}

	crawlResult.SomethingInteresting = crawlResult.One.SomethingInteresting ||
		crawlResult.Two.SomethingInteresting ||
		crawlResult.Three.SomethingInteresting

	return crawlResult, nil
}
