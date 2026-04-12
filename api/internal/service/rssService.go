package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/types"
)

type RssService struct {
	httpClient *http.Client
}

func NewRssService(httpClient *http.Client) *RssService {
	return &RssService{httpClient: httpClient}
}

func (s *RssService) FetchFeed(ctx context.Context, url string, lastRetrieved time.Time) (*types.RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// req.Header.Set("User-Agent", "custom-newsletter")
	
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

	var items []types.RSSItem
	for _, item := range rssFeed.Channel.Items {
		parsedPubDate, err := parseFeedPubDate(item.PubDate)
		if err != nil {
			log.Printf("%v", err)
			continue
		}

		if (parsedPubDate.Before(lastRetrieved)) {
			continue
		}
		
		items = append(items, types.RSSItem{
			Title: html.EscapeString(item.Title),
			Description: html.EscapeString(item.Description),
			Link: item.Link,
			PubDate: item.PubDate,
		})
	}

	rssFeed.Channel.Items = items

	return &rssFeed, nil
}

var possibleRssDateFormats = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	"2006-01-02T15:04:05Z07:00",
    "2006-01-02",
}

func parseFeedPubDate(dateString string) (time.Time, error) {
	s := strings.TrimSpace(dateString)
	for _, format := range possibleRssDateFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("Could not parse date: %q", s)
}