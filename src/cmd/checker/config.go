package main

import (
	"fmt"

	"github.com/Netflix/go-env"
)

type config struct {
	TwoCaptcha struct {
		APIKey string `env:"TWO_CAPTCHA_API_KEY,required=true"`
	}
	ArtifactsDirectory string `env:"ARTIFACTS_DIRECTORY,required=true"`
	RecipientStorage   struct {
		Directory string `env:"RECIPIENT_STORAGE_DIRECTORY,required=true"`
		Limit     uint8  `env:"RECIPIENT_STORAGE_LIMIT,required=true"`
	}
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN,required=true"`
}

func mustLoadConfig() *config {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	return cfg
}

func loadConfig() (*config, error) {
	var c config

	if _, err := env.UnmarshalFromEnviron(&c); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}

	return &c, nil
}
