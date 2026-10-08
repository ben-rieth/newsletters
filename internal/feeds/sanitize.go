package feeds

import (
	"cmp"
	"net/url"
	"slices"

	"github.com/mmcdole/gofeed"
)

const (
	maxFeedItems        = 500
	maxTitleRunes       = 500
	maxDescriptionRunes = 2000
	maxStoredUrlLength  = 2048
)

// Feed content is attacker-controlled and ends up in emails, the web UI, and
// the /link redirect, so only bounded text and http(s) links are kept.
func sanitizeFeed(feed *gofeed.Feed, feedUrl string) *gofeed.Feed {
	base, _ := url.Parse(feedUrl)

	feed.Title = truncateRunes(feed.Title, maxTitleRunes)
	feed.Description = truncateRunes(feed.Description, maxDescriptionRunes)
	feed.Link = safeLink(base, feed.Link)

	items := make([]*gofeed.Item, 0, min(len(feed.Items), maxFeedItems))
	for _, item := range feed.Items {
		link := safeLink(base, item.Link)
		if link == "" {
			continue
		}

		item.Link = link
		item.Title = truncateRunes(item.Title, maxTitleRunes)
		items = append(items, item)
	}

	slices.SortStableFunc(items, func(a, b *gofeed.Item) int {
		return cmp.Compare(publishedUnix(b), publishedUnix(a))
	})
	if len(items) > maxFeedItems {
		items = items[:maxFeedItems]
	}

	feed.Items = items
	return feed
}

func safeLink(base *url.URL, raw string) string {
	if raw == "" || len(raw) > maxStoredUrlLength {
		return ""
	}

	link, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if base != nil {
		link = base.ResolveReference(link)
	}

	if !isWebUrl(link) {
		return ""
	}

	resolved := link.String()
	if len(resolved) > maxStoredUrlLength {
		return ""
	}
	return resolved
}

func isWebUrl(u *url.URL) bool {
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func IsWebUrl(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && isWebUrl(u)
}

func truncateRunes(s string, max int) string {
	count := 0
	for i := range s {
		if count == max {
			return s[:i]
		}
		count++
	}
	return s
}

func publishedUnix(item *gofeed.Item) int64 {
	if item.PublishedParsed == nil {
		return 0
	}
	return item.PublishedParsed.Unix()
}
