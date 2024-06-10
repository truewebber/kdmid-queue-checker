package daemon

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"kdmid-queue-checker/domain/log"
	"kdmid-queue-checker/domain/notification"
)

type NotifierBot struct {
	storage     notification.Storage
	telegramBot *bot.Bot
	logger      log.Logger
}

func NewNotifierBot(botToken string, storage notification.Storage, logger log.Logger) (*NotifierBot, error) {
	notifierBot := &NotifierBot{
		storage: storage,
		logger:  logger,
	}

	if err := notifierBot.registerBot(botToken); err != nil {
		return nil, fmt.Errorf("failed to register bot: %w", err)
	}

	return notifierBot, nil
}

func MustNewNotifierBot(botToken string, storage notification.Storage, logger log.Logger) *NotifierBot {
	notifierBot, err := NewNotifierBot(botToken, storage, logger)
	if err != nil {
		panic(err)
	}

	return notifierBot
}

func (b *NotifierBot) Run(ctx context.Context) error {
	b.telegramBot.Start(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}

func (b *NotifierBot) registerBot(botToken string) error {
	telegramBot, err := bot.New(botToken)
	if err != nil {
		return fmt.Errorf("new bot: %w", err)
	}

	telegramBot.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/register ",
		bot.MatchTypePrefix,
		b.registerHandler,
	)

	telegramBot.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/stop",
		bot.MatchTypeExact,
		b.unregisterHandler,
	)

	telegramBot.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/unregister",
		bot.MatchTypeExact,
		b.unregisterHandler,
	)

	telegramBot.RegisterHandlerRegexp(
		bot.HandlerTypeMessageText,
		regexp.MustCompile(".*"),
		b.defaultHandler,
	)

	b.telegramBot = telegramBot

	return nil
}

func (b *NotifierBot) defaultHandler(ctx context.Context, telegramBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "To use bot please use one of commands below:\n - /register {id} {cd}\n - /stop or /unregister",
	}); err != nil {
		b.logger.Error("send message error", "message", update.Message, "error", err)
	}
}

func (b *NotifierBot) registerHandler(ctx context.Context, telegramBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	chatType := update.Message.Chat.Type
	argsString := strings.TrimPrefix(strings.TrimSpace(update.Message.Text), "/register ")

	if chatType != "private" {
		return
	}

	args := strings.Split(argsString, " ")
	if len(args) != 2 {
		b.logger.Info("invalid amount of register arguments", "message", update.Message)

		return
	}

	r := notification.Recipient{
		TelegramID: chatID,
		ID:         args[0],
		CD:         args[1],
	}

	storageErr := b.storage.Register(ctx, r)

	if errors.Is(storageErr, notification.ErrStorageLimitExceeded) {
		if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Can't register you, my limits exceeded. Try to ask @truewebber, may he can help you.",
		}); err != nil {
			b.logger.Error("send message error", "message", update.Message, "error", err)
		}

		return
	}

	if errors.Is(storageErr, notification.ErrAlreadyExists) {
		if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "One application per user. No more, man.",
		}); err != nil {
			b.logger.Error("send message error", "message", update.Message, "error", err)
		}

		return
	}

	if storageErr != nil {
		b.logger.Error(
			"register storage error",
			"recipient", r,
			"message", update.Message,
			"error", storageErr,
		)

		return
	}

	if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Registered successfully, wait for updates",
	}); err != nil {
		b.logger.Error("send message error", "message", update.Message, "error", err)
	}
}

func (b *NotifierBot) unregisterHandler(ctx context.Context, telegramBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	chatType := update.Message.Chat.Type

	if chatType != "private" {
		return
	}

	r := notification.Recipient{
		TelegramID: chatID,
	}

	storageErr := b.storage.Unregister(ctx, r)

	if errors.Is(storageErr, notification.ErrNotExists) {
		if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "You're not registered yet.",
		}); err != nil {
			b.logger.Error("send message error", "message", update.Message, "error", err)
		}

		return
	}

	if storageErr != nil {
		b.logger.Error(
			"unregister storage error",
			"recipient", r,
			"message", update.Message,
			"error", storageErr,
		)

		return
	}

	if _, err := telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Unregistered successfully",
	}); err != nil {
		b.logger.Error("send message error", "message", update.Message, "error", err)
	}
}
