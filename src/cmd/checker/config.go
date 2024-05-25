package main

import (
	"fmt"

	"github.com/Netflix/go-env"
)

type config struct {
	Application struct {
		ID     string `env:"APPLICATION_ID,required=true"`
		Secret string `env:"APPLICATION_SECRET,required=true"`
	}
	TwoCaptcha struct {
		APIKey string `env:"TWO_CAPTCHA_API_KEY,required=true"`
	}
	ArtifactsDirectory string `env:"ARTIFACTS_DIRECTORY,required=true"`
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
