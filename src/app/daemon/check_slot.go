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
	logger           log.Logger
}

func NewCheckSlot(
	dispatcher page.Dispatcher,
	solver captcha.Solver,
	crawlStorage crawl.Storage,
	recipientStorage notification.Storage,
	logger log.Logger,
) *CheckSlot {
	return &CheckSlot{
		dispatcher:       dispatcher,
		solver:           solver,
		crawlStorage:     crawlStorage,
		recipientStorage: recipientStorage,
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
		if err := c.runSingleCheck(recipient.ID, recipient.CD); err != nil {
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

func (c *CheckSlot) runSingleCheck(applicationID, applicationCD string) error {
	c.logger.Info("start run single check")

	navigator, err := c.dispatcher.NewNavigator(applicationID, applicationCD)
	if err != nil {
		return fmt.Errorf("new navigator: %w", err)
	}

	defer c.logger.CloseWithLog(navigator)

	crawlResult := crawl.Result{
		RanAt: time.Now(),
	}

	crawlResult.One, err = navigator.OpenPageToAuthorize()
	if err != nil {
		return fmt.Errorf("open page to authorize: %w", err)
	}

	code, err := c.solver.Solve(crawlResult.One.Captcha.Image)
	if err != nil {
		return fmt.Errorf("solve captcha: %w", err)
	}

	crawlResult.Two, err = navigator.SubmitAuthorization(code)
	if err != nil {
		return fmt.Errorf("submit authorization: %w", err)
	}

	crawlResult.Three, err = navigator.OpenSlotBookingPage()
	if err != nil {
		return fmt.Errorf("open slot booking page: %w", err)
	}

	crawlResult.SomethingInteresting = crawlResult.One.SomethingInteresting ||
		crawlResult.Two.SomethingInteresting ||
		crawlResult.Three.SomethingInteresting

	if err := c.crawlStorage.Save(crawlResult); err != nil {
		return fmt.Errorf("save crawl result: %w", err)
	}

	c.logger.Info("run single check finished", "something_interesting", crawlResult.SomethingInteresting)

	return nil
}
