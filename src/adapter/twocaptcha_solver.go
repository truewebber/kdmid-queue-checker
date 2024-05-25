package adapter

import (
	"encoding/base64"
	"fmt"

	api2captcha "github.com/2captcha/2captcha-go"

	"kdmid-queue-checker/domain/captcha"
	"kdmid-queue-checker/domain/image"
)

type twoCaptchaSolver struct {
	client   *api2captcha.Client
	numberic int
	maxLen   int
	minLen   int
}

func NewTwoCaptchaSolver(apiKey string) captcha.Solver {
	const (
		pollingInterval = 5
		numberic        = 1
		maxLen          = 6
		minLen          = 6
	)

	client := api2captcha.NewClient(apiKey)
	client.PollingInterval = pollingInterval

	return &twoCaptchaSolver{
		client:   client,
		numberic: numberic,
		maxLen:   maxLen,
		minLen:   minLen,
	}
}

func (t *twoCaptchaSolver) Solve(imageBytes image.PNG) (string, error) {
	normal := api2captcha.Normal{
		Base64:   base64.RawStdEncoding.EncodeToString(imageBytes),
		Numberic: t.numberic,
		MaxLen:   t.maxLen,
		MinLen:   t.minLen,
	}

	code, err := t.client.Solve(normal.ToRequest())
	if err != nil {
		return "", fmt.Errorf("could not solve captcha: %w", err)
	}

	return code, nil
}
