package feeds

import (
	"testing"
	"time"
)

func TestBuildFeedHealth(t *testing.T) {
	lastSuccess := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	before := lastSuccess.Add(-time.Hour)
	after := lastSuccess.Add(time.Hour)

	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	tests := []struct {
		name              string
		disabledUntil     *time.Time
		lastFailure       *FeedFailure
		wantStatus        FeedHealthStatus
		wantDisabledUntil bool
		wantFailureAt     bool
	}{
		{
			name:       "no failure on record is healthy",
			wantStatus: FeedHealthOk,
		},
		{
			name:          "failure older than the last success is healthy",
			lastFailure:   &FeedFailure{OccurredAt: before, Message: "timeout"},
			wantStatus:    FeedHealthOk,
			wantFailureAt: true,
		},
		{
			name:          "failure newer than the last success is failing",
			lastFailure:   &FeedFailure{OccurredAt: after, Message: "timeout"},
			wantStatus:    FeedHealthFailing,
			wantFailureAt: true,
		},
		{
			name:              "an open pause outranks failing",
			disabledUntil:     &future,
			lastFailure:       &FeedFailure{OccurredAt: after, Message: "timeout"},
			wantStatus:        FeedHealthDisabled,
			wantDisabledUntil: true,
			wantFailureAt:     true,
		},
		{
			name:              "a pause that has lapsed is not reported",
			disabledUntil:     &past,
			lastFailure:       &FeedFailure{OccurredAt: after, Message: "timeout"},
			wantStatus:        FeedHealthFailing,
			wantDisabledUntil: false,
			wantFailureAt:     true,
		},
		{
			name:              "an open pause with no failure row is still disabled",
			disabledUntil:     &future,
			wantStatus:        FeedHealthDisabled,
			wantDisabledUntil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildFeedHealth(tt.disabledUntil, lastSuccess, tt.lastFailure)

			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}

			if (got.DisabledUntil != nil) != tt.wantDisabledUntil {
				t.Errorf("DisabledUntil set = %v, want %v", got.DisabledUntil != nil, tt.wantDisabledUntil)
			}

			if (got.LastFailureAt != nil) != tt.wantFailureAt {
				t.Errorf("LastFailureAt set = %v, want %v", got.LastFailureAt != nil, tt.wantFailureAt)
			}

			if !got.LastSuccessAt.Equal(lastSuccess) {
				t.Errorf("LastSuccessAt = %s, want %s", got.LastSuccessAt, lastSuccess)
			}

			if tt.lastFailure != nil && got.LastFailureMessage != tt.lastFailure.Message {
				t.Errorf("LastFailureMessage = %q, want %q", got.LastFailureMessage, tt.lastFailure.Message)
			}
		})
	}
}

// The cutoff bounds the failure lookups that feed BuildFeedHealth, so it has to
// stay old enough to still cover a feed sitting in its longest possible pause.
func TestFailureDisplayCutoffOutlastsTheLongestPause(t *testing.T) {
	age := time.Since(FailureDisplayCutoff())

	if age < maxDisableDuration {
		t.Errorf("cutoff is %s old, want at least the %s max pause", age, maxDisableDuration)
	}

	if age > failureRetention {
		t.Errorf("cutoff is %s old, past the %s retention window that prunes those rows", age, failureRetention)
	}
}
