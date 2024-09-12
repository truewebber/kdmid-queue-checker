package query

import (
	"context"
	"fmt"
	"time"

	crawldomain "github.com/truewebber/kdmid-queue-checker/domain/crawl"
	"github.com/truewebber/kdmid-queue-checker/domain/image"
)

type ListCrawlsHandler struct {
	crawlStorage crawldomain.Storage
}

func NewListCrawlsHandler(crawlStorage crawldomain.Storage) *ListCrawlsHandler {
	return &ListCrawlsHandler{
		crawlStorage: crawlStorage,
	}
}

type Crawl struct {
	Screenshots          []image.PNG
	Captch               image.PNG
	CrawledAt            time.Time
	Err                  error
	SomethingInteresting bool
}

func (h *ListCrawlsHandler) Handle(ctx context.Context, userID int64, date time.Time) ([]Crawl, error) {
	results, err := h.crawlStorage.ListResults(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("list crawls: %w", err)
	}

	return h.castCrawls(results), nil
}

func (h *ListCrawlsHandler) castCrawls(domainCrawls []crawldomain.Result) []Crawl {
	crawls := make([]Crawl, 0, len(domainCrawls))

	for _, domainCrawl := range domainCrawls {
		crawl := Crawl{
			Screenshots: []image.PNG{
				domainCrawl.One.Screenshot,
				domainCrawl.Two.Screenshot,
				domainCrawl.Three.Screenshot,
			},
			Captch:               domainCrawl.One.Captcha.Image,
			CrawledAt:            domainCrawl.RanAt,
			Err:                  domainCrawl.Err,
			SomethingInteresting: domainCrawl.SomethingInteresting,
		}

		crawls = append(crawls, crawl)
	}

	return crawls
}
