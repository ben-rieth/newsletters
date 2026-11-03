package feeds

import (
	"context"
	"errors"
	"testing"
)

func TestIsSafeFeedUrlRequiresHTTPS(t *testing.T) {
	err := IsSafeFeedUrl(context.Background(), "http://example.com/feed")
	if !errors.Is(err, httpsError) {
		t.Fatalf("got %v, want httpsError", err)
	}
}

// IP literals skip DNS entirely, so these exercise the block list without
// touching the network.
func TestIsSafeFeedUrlRejectsInternalAddresses(t *testing.T) {
	for _, rawUrl := range []string{
		"https://127.0.0.1/feed",
		"https://100.64.0.1/feed",
		"https://[::1]/feed",
		"https://[64:ff9b::7f00:1]/feed",
	} {
		if err := IsSafeFeedUrl(context.Background(), rawUrl); !errors.Is(err, invalidIPError) {
			t.Errorf("IsSafeFeedUrl(%q) = %v, want invalidIPError", rawUrl, err)
		}
	}
}

func TestNormalizeItemUrlCollapsesCosmeticDifferences(t *testing.T) {
	want := "https://example.com/posts/hello?p=1"
	for _, rawUrl := range []string{
		"https://example.com/posts/hello?p=1",
		"http://example.com/posts/hello?p=1",
		"https://Example.COM:443/posts/hello/?p=1",
		"https://example.com/posts/hello?utm_source=rss&p=1&fbclid=abc",
	} {
		if got := normalizeItemUrl(rawUrl); got != want {
			t.Errorf("normalizeItemUrl(%q) = %q, want %q", rawUrl, got, want)
		}
	}
}

func TestNormalizeItemUrlKeepsDistinguishingParts(t *testing.T) {
	for _, pair := range [][2]string{
		{"https://example.com/changelog#v1", "https://example.com/changelog#v2"},
		{"https://example.com/?p=1", "https://example.com/?p=2"},
		{"https://example.com/Post", "https://example.com/post"},
	} {
		if normalizeItemUrl(pair[0]) == normalizeItemUrl(pair[1]) {
			t.Errorf("%q and %q normalized to the same URL", pair[0], pair[1])
		}
	}
}

func TestIsSafeFeedUrlHonoursContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := IsSafeFeedUrl(ctx, "https://feeds.example.com/feed")
	if !errors.Is(err, hostResolutionError) {
		t.Fatalf("got %v, want hostResolutionError from the cancelled lookup", err)
	}
}

func TestIsSafeFeedUrlRejectsNonHTTPSPorts(t *testing.T) {
	for _, rawUrl := range []string{
		"https://1.1.1.1:22/feed",
		"https://1.1.1.1:80/feed",
		"https://1.1.1.1:3306/feed",
	} {
		if err := IsSafeFeedUrl(context.Background(), rawUrl); !errors.Is(err, portError) {
			t.Errorf("IsSafeFeedUrl(%q) = %v, want portError", rawUrl, err)
		}
	}
}

func TestIsSafeFeedUrlAllowsHTTPSPorts(t *testing.T) {
	for _, rawUrl := range []string{
		"https://1.1.1.1/feed",
		"https://1.1.1.1:443/feed",
		"https://1.1.1.1:8443/feed",
	} {
		if err := IsSafeFeedUrl(context.Background(), rawUrl); err != nil {
			t.Errorf("IsSafeFeedUrl(%q) = %v, want nil", rawUrl, err)
		}
	}
}
