package crawl

import (
	"context"
	"time"

	"kdmid-queue-checker/domain/page"
)

type Result struct {
	One, Two, Three      page.Stat
	RanAt                time.Time
	SomethingInteresting bool
}

type Storage interface {
	Save(ctx context.Context, userID int64, result *Result) error
	ListUsers(context.Context) ([]int64, error)
	ListResults(ctx context.Context, userID int64, date time.Time) ([]Result, error)
}
