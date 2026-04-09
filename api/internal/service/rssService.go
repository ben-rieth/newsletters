package service

import (
	"context"
	"encoding/xml"
	"html"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/types"
)

type RssService struct {
	httpClient *http.Client
}

func NewRssService(httpClient *http.Client) *RssService {
	return &RssService{httpClient: httpClient}
}

func (s *RssService) FetchFeed(ctx context.Context, url string) (*types.RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "custom-newsletter")
	
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var rssFeed types.RSSFeed
	if err := xml.NewDecoder(res.Body).Decode(&rssFeed); err != nil {
		return nil, err
	}

	rssFeed.Channel.Title = html.EscapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.EscapeString(rssFeed.Channel.Description)

	for _, item := range rssFeed.Channel.Items {
		item.Title = html.EscapeString(item.Title)
		item.Description = html.EscapeString(item.Description)
	}

	return &rssFeed, nil
}