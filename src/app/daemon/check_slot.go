package daemon

import (
	"context"
	"fmt"
	"time"

	"kdmid-queue-checker/domain/captcha"
	"kdmid-queue-checker/domain/crawl"
	"kdmid-queue-checker/domain/log"
	"kdmid-queue-checker/domain/page"
)

type CheckSlot struct {
	dispatcher page.Dispatcher
	solver     captcha.Solver
	storage    crawl.Storage
	logger     log.Logger
}

func NewCheckSlot(
	dispatcher page.Dispatcher,
	solver captcha.Solver,
	storage crawl.Storage,
	logger log.Logger,
) *CheckSlot {
	return &CheckSlot{
		dispatcher: dispatcher,
		solver:     solver,
		storage:    storage,
		logger:     logger,
	}
}

func (c *CheckSlot) Handle(ctx context.Context, applicationID, applicationCD string) error {
	const everyFiveMinutes = 5 * time.Minute

	t := time.NewTimer(0)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := c.runSingleCheck(applicationID, applicationCD); err != nil {
				c.logger.Error("check slot failed", "err", err)
			}

			t.Reset(everyFiveMinutes)
		case <-ctx.Done():
			return nil
		}
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

	if err := c.storage.Save(crawlResult); err != nil {
		return fmt.Errorf("save crawl result: %w", err)
	}

	c.logger.Info("run single check finished", "something_interesting", crawlResult.SomethingInteresting)

	return nil
}
