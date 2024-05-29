package notification

import (
	"context"
	"fmt"
	"time"
)

type Notification struct {
	Images               []PNG
	CrawledAt            time.Time
	Error                error
	SomethingInteresting bool
}

type PNG []byte

type Notifier interface {
	Notify(context.Context, *Notification, *Recipient) error
}

type Recipient struct {
	TelegramID int64
	ID, CD     string
}

var (
	ErrStorageLimitExceeded = fmt.Errorf("storage limit exceeded")
	ErrAlreadyExists        = fmt.Errorf("recipient already exists")
	ErrNotExists            = fmt.Errorf("recipient not exists")
)

type Storage interface {
	Register(Recipient) error
	Unregister(Recipient) error
	List() ([]Recipient, error)
}
