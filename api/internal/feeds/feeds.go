package feeds

import (
	"time"
)

type Feed struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Url string `json:"url"`
	NewsletterId string `json:"newsletterId"`
	LastRetrievedAt *time.Time `json:"lastRetrievedAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BaseFeed struct {
	Id string
	Name string
	URL string
	LastRetrievedAt time.Time
}