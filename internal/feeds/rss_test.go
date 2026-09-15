package feeds

import (
	"errors"
	"fmt"
	"testing"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/utils"
)

// These strings are what a subscriber reads in the failed-feeds section of their
// issue, so an unrecognised error must not leak a raw Go error into the email.
func TestDescribeFetchError(t *testing.T) {
	fetchErr := &FetchError{
		Kind:       db.FeedFetchFailureKindHttpStatus,
		StatusCode: 503,
		Message:    "Feed returned HTTP 503",
		err:        utils.UserError,
	}

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "paused feed", err: ErrFeedDisabled, want: "Paused after repeated failures"},
		{
			name: "wrapped paused feed",
			err:  fmt.Errorf("fetching feed: %w", ErrFeedDisabled),
			want: "Paused after repeated failures",
		},
		{name: "fetch error uses its own message", err: fetchErr, want: "Feed returned HTTP 503"},
		{
			name: "wrapped fetch error still uses its message",
			err:  fmt.Errorf("fetching feed: %w", fetchErr),
			want: "Feed returned HTTP 503",
		},
		{name: "unrecognised error falls back", err: errors.New("dial tcp: i/o timeout"), want: "Could not be retrieved"},
		{name: "nil falls back", err: nil, want: "Could not be retrieved"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DescribeFetchError(tt.err); got != tt.want {
				t.Errorf("DescribeFetchError = %q, want %q", got, tt.want)
			}
		})
	}
}

// Callers branch on UserError vs SystemError to decide whether the fetch failure
// is worth reporting as the user's fault, so the wrapped cause has to survive.
func TestFetchErrorUnwrapsItsCause(t *testing.T) {
	userErr := &FetchError{Message: "Feed URL is not allowed", err: utils.UserError}

	if !errors.Is(userErr, utils.UserError) {
		t.Error("FetchError wrapping UserError did not match it")
	}

	if errors.Is(userErr, utils.SystemError) {
		t.Error("FetchError wrapping UserError matched SystemError")
	}

	if got := userErr.Error(); got != "Feed URL is not allowed" {
		t.Errorf("Error() = %q, want the message", got)
	}
}

func TestErrFeedDisabledIsAUserError(t *testing.T) {
	if !errors.Is(ErrFeedDisabled, utils.UserError) {
		t.Error("ErrFeedDisabled is not a UserError, so a paused feed would be reported as a server fault")
	}
}
