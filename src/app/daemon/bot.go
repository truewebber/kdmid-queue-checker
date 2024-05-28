package daemon

import (
	"context"
	"fmt"
)

type NotifierBot struct {
}

func NewNotifierBot(telegramBotToken, notifierBotDirectory string) *NotifierBot {
	return &NotifierBot{}
}

func (b *NotifierBot) Run(ctx context.Context) error {
	return fmt.Errorf("not implemented yet")
}
