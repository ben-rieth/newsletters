package feeds

import (
	"time"

	"github.com/ben-rieth/newsletter-api/internal/db"
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
	GlobalFeedId string
	NewsletterFeedId string
	Name string
	URL string
	LastRetrievedAt time.Time
}

type FeedFilter struct {
	Id string `json:"id"`
	Field db.FilterField `json:"field"`
	Operator db.FilterOperator `json:"operator"`
	Pattern string `json:"pattern"`
}