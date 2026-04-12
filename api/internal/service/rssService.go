package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type RssService struct {
	httpClient *http.Client
}

func NewRssService(httpClient *http.Client) *RssService {
	return &RssService{httpClient: httpClient}
}

func (s *RssService) FetchFeed(ctx context.Context, url string, lastRetrieved time.Time) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse feed %s: %w", url, err)
	}

	finalItems := make([]*gofeed.Item, 0)
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}

		publishedAt := *item.PublishedParsed

		if publishedAt.Before(lastRetrieved) {
			continue
		}

		finalItems = append(finalItems, item)
	}

	if len(finalItems) == 0 {
		return nil, nil
	}

	feed.Items = finalItems

	return feed, nil
}
