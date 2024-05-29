package daemon

import (
	"context"
	"fmt"
	"time"

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
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			c.runAllRecipients(ctx)
		}
	}
}

const everyFiveMinutes = 5 * time.Minute

func (c *CheckSlot) runAllRecipients(ctx context.Context) {
	t := time.NewTimer(everyFiveMinutes)
	defer t.Stop()

	recipients, err := c.recipientStorage.List()
	if err != nil {
		c.logger.Error("list recipients failed", "err", err)

		<-t.C
	}

	for _, recipient := range recipients {
		if err := c.runSingleCheck(ctx, &recipient); err != nil {
			c.logger.Error("check slot failed", "err", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		t.Reset(everyFiveMinutes)
	}
}

func (c *CheckSlot) runSingleCheck(
	ctx context.Context,
	recipient *notification.Recipient,
) error {
	c.logger.Info("start run single check")

	crawlResult, crawlErr := c.crawl(recipient.ID, recipient.CD)
	if crawlErr != nil {
		if notifyErr := c.notify(ctx, crawlResult, crawlErr, recipient); notifyErr != nil {
			return fmt.Errorf("notify failed: %w: crawl failed: %w", notifyErr, crawlErr)
		}

		return fmt.Errorf("crawl failed: %w", crawlErr)
	}

	if saveErr := c.crawlStorage.Save(crawlResult); saveErr != nil {
		if notifyErr := c.notify(ctx, crawlResult, saveErr, recipient); notifyErr != nil {
			return fmt.Errorf("notify failed: %w: save crawl failed: %w", notifyErr, saveErr)
		}

		return fmt.Errorf("save crawl result: %w", saveErr)
	}

	if crawlResult.SomethingInteresting {
		if notifyErr := c.notify(ctx, crawlResult, nil, recipient); notifyErr != nil {
			return fmt.Errorf("notify failed: %w", notifyErr)
		}
	}

	c.logger.Info("run single check finished", "something_interesting", crawlResult.SomethingInteresting)

	return nil
}

func (c *CheckSlot) notify(
	ctx context.Context,
	result *crawl.Result,
	crawlErr error,
	recipient *notification.Recipient,
) error {
	n := c.buildNotification(result, crawlErr)

	if err := c.notifier.Notify(ctx, n, recipient); err != nil {
		c.logger.Error("notify failed", "err", err)
	}

	return nil
}

func (c *CheckSlot) buildNotification(result *crawl.Result, err error) *notification.Notification {
	return &notification.Notification{
		Images:               c.buildNotificationImages(result),
		CrawledAt:            result.RanAt,
		Error:                err,
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
		return crawlResult, fmt.Errorf("open page to authorize: %w", err)
	}

	code, err := c.solver.Solve(crawlResult.One.Captcha.Image)
	if err != nil {
		return crawlResult, fmt.Errorf("solve captcha: %w", err)
	}

	crawlResult.Two, err = navigator.SubmitAuthorization(code)
	if err != nil {
		return crawlResult, fmt.Errorf("submit authorization: %w", err)
	}

	crawlResult.Three, err = navigator.OpenSlotBookingPage()
	if err != nil {
		return crawlResult, fmt.Errorf("open slot booking page: %w", err)
	}

	crawlResult.SomethingInteresting = crawlResult.One.SomethingInteresting ||
		crawlResult.Two.SomethingInteresting ||
		crawlResult.Three.SomethingInteresting

	return crawlResult, nil
}
