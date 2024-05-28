package adapter

import "kdmid-queue-checker/domain/notification"

type telegramNotifier struct {
}

func NewTelegramNotifier() notification.Notifier {
	return &telegramNotifier{}
}

func (n *telegramNotifier) Notify(notification notification.Notification, recipient notification.Recipient) error {
	return nil
}
