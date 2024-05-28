package notification

import "time"

type Notification struct {
	Images               []PNG
	Text                 string
	CrawledAt            time.Time
	Error                error
	SomethingInteresting bool
}

type PNG []byte

type Notifier interface {
	Notify(Notification, Recipient) error
}

type Recipient struct {
	TelegramID int64
}

type Storage interface {
	Register(Recipient) error
}
