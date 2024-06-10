package crawl

import (
	"time"

	"kdmid-queue-checker/domain/page"
)

type Result struct {
	One, Two, Three      page.Stat
	RanAt                time.Time
	SomethingInteresting bool
}

type Storage interface {
	Save(userID int64, result *Result) error
	ListUsers() ([]int64, error)
	ListResults(userID int64, date time.Time) ([]Result, error)
}
