package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"kdmid-queue-checker/domain/log"
	"kdmid-queue-checker/service"
)

func main() {
	cfg := mustLoadConfig()
	logger := log.NewZapWrapper()
	ctx := contextClosableOnSignals(syscall.SIGINT, syscall.SIGTERM)

	if err := run(ctx, cfg, logger); err != nil {
		panic(err)
	}

	if err := logger.Close(); err != nil {
		panic(err)
	}
}

func run(ctx context.Context, cfg *config, logger log.Logger) error {
	appConfig := &service.Config{
		TwoCaptchaAPIKey:   cfg.TwoCaptcha.APIKey,
		ArtifactsDirectory: cfg.ArtifactsDirectory,
	}

	app := service.NewApplication(appConfig, logger)

	if err := app.Daemon.CheckSlot.Handle(ctx, cfg.Application.ID, cfg.Application.Secret); err != nil {
		return fmt.Errorf("handle daemon check slot: %w", err)
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
