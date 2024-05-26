package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"kdmid-queue-checker/adapter"
	"kdmid-queue-checker/domain/captcha"
	"kdmid-queue-checker/domain/crawl"
	"kdmid-queue-checker/domain/log"
	"kdmid-queue-checker/domain/page"
)

func main() {
	cfg := mustLoadConfig()
	logger := log.NewZapWrapper()

	dispatcher := adapter.MustNewChromeDispatcher()
	solver := adapter.NewTwoCaptchaSolver(cfg.TwoCaptcha.APIKey)
	storage := adapter.MustNewFileSystemCrawlStorage(cfg.ArtifactsDirectory, logger)

	if err := run(cfg.Application.ID, cfg.Application.Secret, dispatcher, solver, storage, logger); err != nil {
		panic(err)
	}

	if err := dispatcher.Close(); err != nil {
		panic(err)
	}

	if err := logger.Close(); err != nil {
		panic(err)
	}
}

func run(
	id, cd string,
	dispatcher page.Dispatcher,
	solver captcha.Solver,
	storage crawl.Storage,
	logger log.Logger,
) error {
	navigator, err := dispatcher.NewNavigator(id, cd)
	if err != nil {
		return fmt.Errorf("new navigator: %w", err)
	}

	defer logger.CloseWithLog(navigator)

	crawlResult := crawl.Result{
		RanAt: time.Now(),
	}

	crawlResult.One, err = navigator.OpenPageToAuthorize()
	if err != nil {
		return fmt.Errorf("open page to authorize: %w", err)
	}

	code, err := solver.Solve(crawlResult.One.Captcha.Image)
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

	if err := storage.Save(crawlResult); err != nil {
		return fmt.Errorf("save crawl result: %w", err)
	}

	return nil
}

func contextClosableOnSignals(sig ...os.Signal) context.Context {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, sig...)

	ctx, cancel := context.WithCancel(context.Background())

	go func(signals <-chan os.Signal) {
		<-signals

		cancel()
	}(signals)

	return ctx
}
