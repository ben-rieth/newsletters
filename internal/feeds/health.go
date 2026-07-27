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

func BuildFeedHealth(
	disabledUntil *time.Time,
	lastSuccessAt time.Time,
	lastFailure *FeedFailure,
) FeedHealth {
	health := FeedHealth{
		Status:        FeedHealthOk,
		DisabledUntil: disabledUntil,
		LastSuccessAt: lastSuccessAt,
	}

	if lastFailure != nil {
		health.LastFailureAt = &lastFailure.OccurredAt
		health.LastFailureMessage = lastFailure.Message

		if lastFailure.OccurredAt.After(lastSuccessAt) {
			health.Status = FeedHealthFailing
		}
	}

	if disabledUntil != nil && disabledUntil.After(time.Now()) {
		health.Status = FeedHealthDisabled
	}

	return health
}
