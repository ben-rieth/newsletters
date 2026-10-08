package feeds

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/utils"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/mmcdole/gofeed"
)

type FeedView struct {
	Title   string
	HtmlURL string
	Items   []FeedItemView
}

type FeedItemView struct {
	ItemID      string
	Title       string
	URL         string
	PublishDate time.Time
	TrackingURL string
}

type FeedMetaData struct {
	Id          string
	Title       string
	URL         string
	HtmlURL     string
	Description string
}

type FetchFeedResult struct {
	Feed        *gofeed.Feed
	FinalUrl    string
	OriginalUrl string
	RetrievedAt time.Time
}

type FetchError struct {
	Kind       db.FeedFetchFailureKind
	StatusCode int
	Message    string
	err        error
}

func (e *FetchError) Error() string {
	return e.Message
}

func (e *FetchError) Unwrap() error {
	return e.err
}

func DescribeFetchError(err error) string {
	if errors.Is(err, ErrFeedDisabled) {
		return "Paused after repeated failures"
	}

	var fetchErr *FetchError
	if errors.As(err, &fetchErr) {
		return fetchErr.Message
	}

	return "Could not be retrieved"
}

type RssService struct {
	httpClient *http.Client
}

func NewRssService() *RssService {
	client := newSafeFeedClient()
	return &RssService{httpClient: client}
}

func (s *RssService) FetchFeed(ctx context.Context, url string) (*FetchFeedResult, error) {
	// Unresolvable and internal hosts share one message so the response can't
	// be used to discover which hostnames point inside the network.
	err := IsSafeFeedUrl(ctx, url)
	if errors.Is(err, hostResolutionError) {
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindTransport,
			Message: "Feed could not be reached",
			err:     utils.UserError,
		}
	} else if errors.Is(err, invalidIPError) {
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindUnsafeUrl,
			Message: "Feed could not be reached",
			err:     utils.UserError,
		}
	} else if err != nil {
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindUnsafeUrl,
			Message: "Feed URL is not allowed",
			err:     utils.UserError,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindTransport,
			Message: "Feed URL could not be requested",
			err:     utils.UserError,
		}
	}

	res, err := s.httpClient.Do(req)
	if errors.Is(err, errInsecureRedirect) {
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindUnsafeUrl,
			Message: errInsecureRedirect.Error(),
			err:     utils.UserError,
		}
	}
	if err != nil {
		wideLog.AddErrorField(ctx, err)
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindTransport,
			Message: "Feed could not be reached",
			err:     utils.SystemError,
		}
	}

	if res != nil {
		defer res.Body.Close()
	}

	wideLog.AddLogField(ctx, "feedStatusCode", res.StatusCode)
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return nil, &FetchError{
			Kind:       db.FeedFetchFailureKindHttpStatus,
			StatusCode: res.StatusCode,
			Message:    fmt.Sprintf("Feed returned HTTP %d", res.StatusCode),
			err:        utils.UserError,
		}
	}

	finalUrl := res.Request.URL.String()

	feed, err := parseFeedBody(res.Body)
	if err != nil {
		wideLog.AddErrorField(ctx, err)
		return nil, err
	}

	retrievedAt := time.Now()

	return &FetchFeedResult{
		Feed:        feed,
		FinalUrl:    finalUrl,
		OriginalUrl: url,
		RetrievedAt: retrievedAt,
	}, nil
}

// The cap stops a hostile or broken URL from streaming until the client
// timeout and holding all of it in memory.
const maxFeedBytes = 10 << 20

func parseFeedBody(body io.Reader) (*gofeed.Feed, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxFeedBytes+1))
	if err != nil {
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindTransport,
			Message: "Feed could not be reached",
			err:     utils.SystemError,
		}
	}

	if len(data) > maxFeedBytes {
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindParse,
			Message: fmt.Sprintf("Feed is larger than %d MB", maxFeedBytes>>20),
			err:     utils.UserError,
		}
	}

	feed, err := gofeed.NewParser().Parse(bytes.NewReader(data))
	if err != nil {
		return nil, &FetchError{
			Kind:    db.FeedFetchFailureKindParse,
			Message: "Feed response was not valid RSS or Atom",
			err:     utils.SystemError,
		}
	}

	return feed, nil
}
