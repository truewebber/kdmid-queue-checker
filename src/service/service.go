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
	crawlStorage := adapter.MustNewFileSystemCrawlStorage(cfg.ArtifactsDirectory, logger)
	recipientStorage := adapter.MustNewRecipientStorageFs(
		cfg.RecipientStorage.Directory, cfg.RecipientStorage.Limit, logger,
	)

	return &app.Application{
		Daemon: app.Daemon{
			CheckSlot: daemon.NewCheckSlot(dispatcher, solver, crawlStorage, recipientStorage, nil, logger),
			Bot:       daemon.MustNewNotifierBot(cfg.TelegramBotToken, recipientStorage, logger),
		},
	}
}

type Config struct {
	TwoCaptchaAPIKey   string
	ArtifactsDirectory string
	TelegramBotToken   string
	RecipientStorage   RecipientStorage
}

type RecipientStorage struct {
	Directory string
	Limit     uint8
}
