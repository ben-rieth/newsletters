package feeds

import (
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
)

type Feed struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Url             string     `json:"url"`
	NewsletterId    string     `json:"newsletterId"`
	LastRetrievedAt *time.Time `json:"lastRetrievedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type BaseFeed struct {
	GlobalFeedId     string
	NewsletterFeedId string
	Name             string
	URL              string
	LastRetrievedAt  time.Time
}

type FeedFilter struct {
	Id       string            `json:"id"`
	Field    db.FilterField    `json:"field"`
	Operator db.FilterOperator `json:"operator"`
	Pattern  string            `json:"pattern"`
}

type ExportableFeed struct {
	ID       string       `json:"id"`
	GlobalID string       `json:"globalId"`
	Name     string       `json:"name"`
	Alias    string       `json:"alias"`
	URL      string       `json:"url"`
	Filters  []FeedFilter `json:"filters"`
}
