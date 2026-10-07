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
	HtmlURL          string
	LastRetrievedAt  time.Time
}

type FeedFilter struct {
	Id       string            `json:"id"`
	Field    db.FilterField    `json:"field" enum:"title,url"`
	Operator db.FilterOperator `json:"operator" enum:"contains,does_not_contain"`
	Pattern  string            `json:"pattern"`
}

type ExportableFeed struct {
	ID       string             `json:"id" required:"false"`
	GlobalID string             `json:"globalId" required:"false"`
	Name     string             `json:"name" required:"false"`
	Alias    string             `json:"alias" required:"false"`
	URL      string             `json:"url" minLength:"1" maxLength:"2048"`
	Status   string             `json:"status" enum:"active,inactive" required:"false"`
	Filters  []ExportableFilter `json:"filters" required:"false"`
}

type ExportableFilter struct {
	Id       string            `json:"id" required:"false"`
	Field    db.FilterField    `json:"field" enum:"title,url"`
	Operator db.FilterOperator `json:"operator" enum:"contains,does_not_contain"`
	Pattern  string            `json:"pattern"`
}
