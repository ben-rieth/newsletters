package feeds

import (
	"testing"
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/mmcdole/gofeed"
)

func item(title, link string, published *time.Time) *gofeed.Item {
	return &gofeed.Item{Title: title, Link: link, PublishedParsed: published}
}

func at(t time.Time) *time.Time {
	return &t
}

func TestPruneFeedItemsBeforeTime(t *testing.T) {
	threshold := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)

	feed := &gofeed.Feed{Items: []*gofeed.Item{
		item("older", "https://example.com/older", at(threshold.Add(-time.Minute))),
		item("exactly at threshold", "https://example.com/exact", at(threshold)),
		item("newer", "https://example.com/newer", at(threshold.Add(time.Minute))),
		item("undated", "https://example.com/undated", nil),
	}}

	got := pruneFeedItemsBeforeTime(feed, threshold)

	var titles []string
	for _, kept := range got.Items {
		titles = append(titles, kept.Title)
	}

	// The threshold is the last time the feed was read, so an item stamped exactly
	// then has not been delivered yet and has to survive.
	want := []string{"exactly at threshold", "newer"}

	if len(titles) != len(want) {
		t.Fatalf("kept %v, want %v", titles, want)
	}

	for i, title := range want {
		if titles[i] != title {
			t.Errorf("kept[%d] = %q, want %q", i, titles[i], title)
		}
	}
}

func TestPruneFeedItemsBeforeTimeDropsEverything(t *testing.T) {
	threshold := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)

	feed := &gofeed.Feed{Items: []*gofeed.Item{
		item("older", "https://example.com/older", at(threshold.Add(-time.Hour))),
		item("undated", "https://example.com/undated", nil),
	}}

	got := pruneFeedItemsBeforeTime(feed, threshold)

	if got.Items == nil {
		t.Error("Items = nil, want an empty slice")
	}

	if len(got.Items) != 0 {
		t.Errorf("kept %d items, want 0", len(got.Items))
	}
}

func TestBuildFeedItemParamsFromGoFeed(t *testing.T) {
	published := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	retrievedAt := time.Date(2026, time.March, 11, 9, 0, 0, 0, time.UTC)

	feed := &gofeed.Feed{Items: []*gofeed.Item{
		item("dated", "https://example.com/dated", at(published)),
		// An undated item cannot be placed in any issue's window, so it is dropped
		// rather than stored with a zero publish date.
		item("undated", "https://example.com/undated", nil),
	}}

	got := buildFeedItemParamsFromGoFeed(feed, "feed-1", retrievedAt)

	if len(got) != 1 {
		t.Fatalf("got %d params, want 1", len(got))
	}

	want := db.SaveFeedItemDetailsParams{
		Title:       "dated",
		Url:         "https://example.com/dated",
		PublishDate: published,
		FeedID:      "feed-1",
		RetrievedAt: retrievedAt,
	}

	if got[0] != want {
		t.Errorf("params = %+v, want %+v", got[0], want)
	}
}

func sourceByUrl(params []db.SaveFeedUrlsParams) map[string]db.FeedUrlSource {
	sources := make(map[string]db.FeedUrlSource, len(params))
	for _, param := range params {
		sources[param.Url] = param.Source
	}

	return sources
}

func TestBuildUrlList(t *testing.T) {
	tests := []struct {
		name        string
		originalUrl string
		finalUrl    string
		want        map[string]db.FeedUrlSource
	}{
		{
			name:        "a redirect records both urls",
			originalUrl: "https://example.com/feed",
			finalUrl:    "https://example.com/feed.xml",
			want: map[string]db.FeedUrlSource{
				"https://example.com/feed":     db.FeedUrlSourceUserSubmitted,
				"https://example.com/feed.xml": db.FeedUrlSourceCanonical,
			},
		},
		{
			name:        "no redirect records the url once as canonical",
			originalUrl: "https://example.com/feed",
			finalUrl:    "https://example.com/feed",
			want: map[string]db.FeedUrlSource{
				"https://example.com/feed": db.FeedUrlSourceCanonical,
			},
		},
		{
			name:     "an empty original url is skipped",
			finalUrl: "https://example.com/feed.xml",
			want: map[string]db.FeedUrlSource{
				"https://example.com/feed.xml": db.FeedUrlSourceCanonical,
			},
		},
		{
			name:        "an empty final url is skipped",
			originalUrl: "https://example.com/feed",
			want: map[string]db.FeedUrlSource{
				"https://example.com/feed": db.FeedUrlSourceUserSubmitted,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildUrlList(&FetchFeedResult{
				OriginalUrl: tt.originalUrl,
				FinalUrl:    tt.finalUrl,
			}, "feed-1")

			if len(got) != len(tt.want) {
				t.Fatalf("got %d urls, want %d: %+v", len(got), len(tt.want), got)
			}

			for url, source := range sourceByUrl(got) {
				wantSource, ok := tt.want[url]
				if !ok {
					t.Errorf("unexpected url %q", url)
					continue
				}

				if source != wantSource {
					t.Errorf("source for %q = %q, want %q", url, source, wantSource)
				}
			}

			for _, param := range got {
				if param.FeedID != "feed-1" {
					t.Errorf("FeedID = %q, want feed-1", param.FeedID)
				}
			}
		})
	}
}
