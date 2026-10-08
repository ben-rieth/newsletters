package feeds

import (
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func itemWithLink(link string) *gofeed.Item {
	published := time.Now()
	return &gofeed.Item{Title: "t", Link: link, PublishedParsed: &published}
}

func TestSanitizeFeedLinks(t *testing.T) {
	tests := []struct {
		name string
		link string
		want string
	}{
		{name: "https kept", link: "https://example.com/a", want: "https://example.com/a"},
		{name: "http kept", link: "http://example.com/a", want: "http://example.com/a"},
		{name: "relative resolved against feed", link: "/posts/1", want: "https://blog.example.com/posts/1"},
		{name: "javascript dropped", link: "javascript:alert(1)", want: ""},
		{name: "mixed-case javascript dropped", link: "JavaScript:alert(1)", want: ""},
		{name: "data dropped", link: "data:text/html,<script>alert(1)</script>", want: ""},
		{name: "empty dropped", link: "", want: ""},
		{name: "overlong dropped", link: "https://example.com/" + strings.Repeat("a", maxStoredUrlLength), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed := &gofeed.Feed{Link: tt.link, Items: []*gofeed.Item{itemWithLink(tt.link)}}

			got := sanitizeFeed(feed, "https://blog.example.com/feed.xml")

			if got.Link != tt.want {
				t.Errorf("feed link = %q, want %q", got.Link, tt.want)
			}
			if tt.want == "" && len(got.Items) != 0 {
				t.Errorf("item with link %q was kept", tt.link)
			}
			if tt.want != "" && (len(got.Items) != 1 || got.Items[0].Link != tt.want) {
				t.Errorf("item link not %q: %+v", tt.want, got.Items)
			}
		})
	}
}

func TestSanitizeFeedTruncatesText(t *testing.T) {
	item := itemWithLink("https://example.com/a")
	item.Title = strings.Repeat("é", maxTitleRunes+10)
	feed := &gofeed.Feed{
		Title:       strings.Repeat("é", maxTitleRunes+10),
		Description: strings.Repeat("x", maxDescriptionRunes+10),
		Items:       []*gofeed.Item{item},
	}

	got := sanitizeFeed(feed, "https://example.com/feed")

	if n := len([]rune(got.Title)); n != maxTitleRunes {
		t.Errorf("feed title has %d runes, want %d", n, maxTitleRunes)
	}
	if n := len([]rune(got.Description)); n != maxDescriptionRunes {
		t.Errorf("feed description has %d runes, want %d", n, maxDescriptionRunes)
	}
	if n := len([]rune(got.Items[0].Title)); n != maxTitleRunes {
		t.Errorf("item title has %d runes, want %d", n, maxTitleRunes)
	}
}

func TestSanitizeFeedKeepsNewestItems(t *testing.T) {
	oldest := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	items := make([]*gofeed.Item, 0, maxFeedItems+5)
	for i := range maxFeedItems + 5 {
		published := oldest.Add(time.Duration(i) * time.Hour)
		items = append(items, &gofeed.Item{Link: "https://example.com/a", PublishedParsed: &published})
	}

	got := sanitizeFeed(&gofeed.Feed{Items: items}, "https://example.com/feed")

	if len(got.Items) != maxFeedItems {
		t.Fatalf("kept %d items, want %d", len(got.Items), maxFeedItems)
	}
	for _, item := range got.Items {
		if item.PublishedParsed.Before(oldest.Add(5 * time.Hour)) {
			t.Fatalf("kept an older item (%s) over a newer one", item.PublishedParsed)
		}
	}
}

func TestIsWebUrl(t *testing.T) {
	for raw, want := range map[string]bool{
		"https://example.com/a": true,
		"http://example.com/a":  true,
		"javascript:alert(1)":   false,
		"/relative":             false,
		"https:///no-host":      false,
		"":                      false,
	} {
		if got := IsWebUrl(raw); got != want {
			t.Errorf("IsWebUrl(%q) = %v, want %v", raw, got, want)
		}
	}
}
