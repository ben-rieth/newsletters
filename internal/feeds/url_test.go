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

func TestIsSafeFeedUrlHonoursContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := IsSafeFeedUrl(ctx, "https://feeds.example.com/feed")
	if !errors.Is(err, hostResolutionError) {
		t.Fatalf("got %v, want hostResolutionError from the cancelled lookup", err)
	}
}
