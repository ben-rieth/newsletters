package feeds

import "time"

type FeedHealthStatus string

const (
	FeedHealthOk       FeedHealthStatus = "ok"
	FeedHealthFailing  FeedHealthStatus = "failing"
	FeedHealthDisabled FeedHealthStatus = "disabled"
)

type FeedFailure struct {
	OccurredAt time.Time
	Message    string
}

type FeedHealth struct {
	Status             FeedHealthStatus `json:"status" enum:"ok,failing,disabled"`
	DisabledUntil      *time.Time       `json:"disabledUntil,omitempty"`
	LastSuccessAt      time.Time        `json:"lastSuccessAt"`
	LastFailureAt      *time.Time       `json:"lastFailureAt,omitempty"`
	LastFailureMessage string           `json:"lastFailureMessage,omitempty"`
}

// FailureDisplayCutoff bounds the failure lookups that feed BuildFeedHealth, so a
// feed that recovered long ago stops reporting an ancient failure and the reads
// stop scanning the feed's whole history.
func FailureDisplayCutoff() time.Time {
	return time.Now().Add(-failureDisplayWindow)
}

func BuildFeedHealth(
	disabledUntil *time.Time,
	lastSuccessAt time.Time,
	lastFailure *FeedFailure,
) FeedHealth {
	stillDisabled := disabledUntil != nil && disabledUntil.After(time.Now())

	health := FeedHealth{
		Status:        FeedHealthOk,
		LastSuccessAt: lastSuccessAt,
	}

	if stillDisabled {
		health.DisabledUntil = disabledUntil
	}

	if lastFailure != nil {
		health.LastFailureAt = &lastFailure.OccurredAt
		health.LastFailureMessage = lastFailure.Message

		if lastFailure.OccurredAt.After(lastSuccessAt) {
			health.Status = FeedHealthFailing
		}
	}

	if stillDisabled {
		health.Status = FeedHealthDisabled
	}

	return health
}
