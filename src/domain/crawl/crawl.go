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
	Save(*Result) error
	List(offset, limit int) ([]Result, error)
}
