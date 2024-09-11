package adapter

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"kdmid-queue-checker/domain/notification"
)

type telegramNotifier struct {
	telegramBot *bot.Bot
}

func NewTelegramNotifier(telegramBotToken string) (notification.Notifier, error) {
	telegramBot, err := bot.New(telegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("new bot: %w", err)
	}

	return &telegramNotifier{
		telegramBot: telegramBot,
	}, nil
}

func MustNewTelegramNotifier(telegramBotToken string) notification.Notifier {
	notifier, err := NewTelegramNotifier(telegramBotToken)
	if err != nil {
		panic(err)
	}

	return notifier
}

func (n *telegramNotifier) Notify(
	ctx context.Context,
	notification *notification.Notification,
	recipient *notification.Recipient,
) error {
	if len(notification.Images) > 0 {
		media := make([]models.InputMedia, 0, len(notification.Images))

		for i, image := range notification.Images {
			media = append(media, &models.InputMediaPhoto{
				Media:           fmt.Sprintf("attach://screenshot%d.png", i+1),
				MediaAttachment: bytes.NewReader(image),
			})
		}

		if _, err := n.telegramBot.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
			ChatID: recipient.TelegramID,
			Media:  media,
		}); err != nil {
			return fmt.Errorf("send message: %w", err)
		}
	}

	text := make([]string, 0)

	if notification.Error != nil {
		text = append(text, "Error occurred during checking.")
	}

	if notification.SomethingInteresting {
		text = append(text, "It something interesting was found, time to visit website.")
	}

	text = append(text, "collected at "+notification.CrawledAt.Format(time.DateTime))

	if _, err := n.telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: recipient.TelegramID,
		Text:   strings.Join(text, "\n"),
	}); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
