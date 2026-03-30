package types

import "time"

type Feed struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Url string `json:"url"`
	NewsletterId string `json:"newsletterId"`
	LastRetrievedAt *time.Time `json:"lastRetrievedAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SubmittableFeedFields struct {
	Name string `json:"name" minLength:"1"`
	Url string `json:"url" doc:"Must be a valid HTTPS URL" pattern:"^https:\\/\\/([\\w-]+\\.)+[\\w]{2,}(\\/[\\w\\-.~:/?#\\[\\]@!$&'()*+,;=%]*)?$"`
}

type AddFeedInput struct {
	NewsletterID string `path:"newsletterId"`
	Body SubmittableFeedFields
}

type ListFeedsInput struct {
	NewsletterID string `path:"newsletterId"`
}

type ListFeedsOutput struct {
	Body []Feed
}