package feeds

import (
	"context"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/utils"
	"github.com/mmcdole/gofeed"
)

type FeedView struct {
	Title string
	URL   string
	Items []FeedItemView
}

type FeedItemView struct {
	Title       string
	URL         string
	PublishDate time.Time
}

type FeedMetaData struct {
	Id          string
	Title       string
	URL         string
	Description string
}

type FetchFeedResult struct {
	Feed        *gofeed.Feed
	FinalUrl    string
	OriginalUrl string
	RetrievedAt time.Time
}

type RssService struct {
	httpClient *http.Client
}

func NewRssService() *RssService {
	client := newSafeFeedClient()
	return &RssService{httpClient: client}
}

func (s *RssService) FetchFeed(ctx context.Context, url string) (*FetchFeedResult, error) {
	err := IsSafeFeedUrl(url)
	if err != nil {
		return nil, utils.UserError
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, utils.SystemError
	}

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, utils.SystemError
	}

	if res != nil {
		defer res.Body.Close()
	}

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return nil, utils.UserError
	}

	finalUrl := res.Request.URL.String()

	fp := gofeed.NewParser()
	feed, err := fp.Parse(res.Body)
	if err != nil {
		return nil, utils.SystemError
	}

	retrievedAt := time.Now()

	return &FetchFeedResult{
		Feed:        feed,
		FinalUrl:    finalUrl,
		OriginalUrl: url,
		RetrievedAt: retrievedAt,
	}, nil
}
