package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"kdmid-queue-checker/domain/log"
	"kdmid-queue-checker/port"
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
	appConfig, err := buildAppConfig(cfg)
	if err != nil {
		return fmt.Errorf("build app config: %w", err)
	}

	app := service.NewApplication(appConfig, logger)

	logger.Info("Application configured")

	httpServer := port.NewHTTP(cfg.AppHostPort, app, logger)

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		if err := app.Daemon.CheckSlot.Handle(groupCtx); err != nil {
			return fmt.Errorf("handle daemon check slot: %w", err)
		}

		return nil
	})

	group.Go(func() error {
		if err := app.Daemon.Bot.Run(groupCtx); err != nil {
			return fmt.Errorf("run bot daemon: %w", err)
		}

		return nil
	})

	group.Go(func() error {
		if err := httpServer.Start(groupCtx); err != nil {
			return fmt.Errorf("run http server: %w", err)
		}

		return nil
	})

	if err := group.Wait(); err != nil {
		return fmt.Errorf("group wait: %w", err)
	}

	return nil
}

func buildAppConfig(cfg *config) (*service.Config, error) {
	proxyURL, err := url.Parse(cfg.ProxyURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy url: %w", err)
	}

	return &service.Config{
		TwoCaptchaAPIKey:   cfg.TwoCaptcha.APIKey,
		ArtifactsDirectory: cfg.ArtifactsDirectory,
		TelegramBotToken:   cfg.TelegramBotToken,
		RecipientStorage: service.RecipientStorage{
			Directory: cfg.RecipientStorage.Directory,
			Limit:     cfg.RecipientStorage.Limit,
		},
		ProxyURL: proxyURL,
	}, nil
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
