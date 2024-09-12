package query

import (
	"context"
	"fmt"
	"sort"

	"github.com/truewebber/kdmid-queue-checker/domain/crawl"
	"github.com/truewebber/kdmid-queue-checker/domain/notification"
)

type ListUsersHandler struct {
	notificationStorage notification.Storage
	crawlStorage        crawl.Storage
}

func NewListUsersHandler(notificationStorage notification.Storage, crawlStorage crawl.Storage) *ListUsersHandler {
	return &ListUsersHandler{
		notificationStorage: notificationStorage,
		crawlStorage:        crawlStorage,
	}
}

type User struct {
	TelegramID        int64
	Active, HasCrawls bool
}

func (h *ListUsersHandler) Handle(ctx context.Context) ([]User, error) {
	activeRecipients, err := h.notificationStorage.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recipinets: %w", err)
	}

	usersWithCrawls, err := h.crawlStorage.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users with crawls: %w", err)
	}

	return h.mergeAndCastUsers(activeRecipients, usersWithCrawls), nil
}

func (h *ListUsersHandler) mergeAndCastUsers(
	activeRecipients []notification.Recipient,
	usersWithCrawls []int64,
) []User {
	usersMap := make(map[int64]User, len(activeRecipients))

	for _, recipient := range activeRecipients {
		usersMap[recipient.TelegramID] = User{
			TelegramID: recipient.TelegramID,
			Active:     true,
			HasCrawls:  false,
		}
	}

	for _, telegramID := range usersWithCrawls {
		u, ok := usersMap[telegramID]
		if ok {
			u.HasCrawls = true
			usersMap[telegramID] = u

			continue
		}

		usersMap[telegramID] = User{
			TelegramID: telegramID,
			Active:     false,
			HasCrawls:  true,
		}
	}

	users := make([]User, 0, len(usersMap))

	for _, user := range usersMap {
		users = append(users, user)
	}

	sort.SliceStable(users, func(i, j int) bool {
		return users[i].TelegramID < users[j].TelegramID
	})

	return users
}
