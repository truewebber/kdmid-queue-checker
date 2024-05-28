package service

import (
	"kdmid-queue-checker/adapter"
	"kdmid-queue-checker/app"
	"kdmid-queue-checker/app/daemon"
	"kdmid-queue-checker/domain/log"
)

func NewApplication(cfg *Config, logger log.Logger) *app.Application {
	dispatcher := adapter.MustNewChromeDispatcher()
	solver := adapter.NewTwoCaptchaSolver(cfg.TwoCaptchaAPIKey)
	storage := adapter.MustNewFileSystemCrawlStorage(cfg.ArtifactsDirectory, logger)

	return &app.Application{
		Daemon: app.Daemon{
			CheckSlot: daemon.NewCheckSlot(dispatcher, solver, storage, logger),
			Bot:       daemon.NewNotifierBot(cfg.TelegramBotToken, cfg.NotifierBotDirectory),
		},
	}
}

type Config struct {
	TwoCaptchaAPIKey     string
	ArtifactsDirectory   string
	TelegramBotToken     string
	NotifierBotDirectory string
}
