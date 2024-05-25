package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"kdmid-queue-checker/adapter"
	"kdmid-queue-checker/domain/captcha"
	"kdmid-queue-checker/domain/page"
)

func main() {
	cfg := mustLoadConfig()

	dispatcher := adapter.MustNewChromeDispatcher()
	solver := adapter.NewTwoCaptchaSolver(cfg.TwoCaptcha.APIKey)

	if err := run(cfg.Application.ID, cfg.Application.Secret, dispatcher, solver); err != nil {
		panic(err)
	}

	if err := dispatcher.Close(); err != nil {
		panic(err)
	}
}

func run(id, cd string, dispatcher page.Dispatcher, solver captcha.Solver) error {
	navigator, err := dispatcher.NewNavigator(id, cd)
	if err != nil {
		return fmt.Errorf("new navigator: %w", err)
	}

	defer navigator.Close()

	stat, err := navigator.OpenPageToAuthorize()
	if err != nil {
		return fmt.Errorf("open slot booking page: %w", err)
	}

	fmt.Printf("%v\n", stat)

	code, err := solver.Solve(stat.Captcha.Image)
	if err != nil {
		return fmt.Errorf("open slot booking page: %w", err)
	}

	stat, err = navigator.SubmitAuthorization(code)
	if err != nil {
		return fmt.Errorf("open slot booking page: %w", err)
	}

	fmt.Printf("%v\n", stat)

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
