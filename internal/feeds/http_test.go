package feeds

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
)

func redirectRequest(t *testing.T, rawUrl string) *http.Request {
	t.Helper()
	u, err := url.Parse(rawUrl)
	if err != nil {
		t.Fatalf("could not parse %q", rawUrl)
	}
	return &http.Request{URL: u}
}

func TestCheckFeedRedirectRejectsDowngradeToHTTP(t *testing.T) {
	via := []*http.Request{redirectRequest(t, "https://example.com/feed")}

	err := checkFeedRedirect(redirectRequest(t, "http://example.com/feed"), via)
	if !errors.Is(err, errInsecureRedirect) {
		t.Fatalf("redirect to http was allowed, got %v", err)
	}
}

func TestCheckFeedRedirectAllowsHTTPS(t *testing.T) {
	via := []*http.Request{redirectRequest(t, "https://example.com/feed")}

	if err := checkFeedRedirect(redirectRequest(t, "https://www.example.com/feed"), via); err != nil {
		t.Fatalf("https redirect was blocked: %v", err)
	}
}

func TestCheckFeedRedirectStopsRedirectLoops(t *testing.T) {
	via := make([]*http.Request, maxFeedRedirects)
	for i := range via {
		via[i] = redirectRequest(t, "https://example.com/feed")
	}

	if err := checkFeedRedirect(redirectRequest(t, "https://example.com/feed"), via); err == nil {
		t.Fatal("redirect past the limit was allowed")
	}
}

func TestSafeFeedClientChecksRedirects(t *testing.T) {
	if newSafeFeedClient().CheckRedirect == nil {
		t.Fatal("client has no redirect check, so redirects to http would be followed")
	}
}
