package types

import (
	"encoding/xml"
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

type UpdateFeedInput struct {
	NewsletterID string `path:"newsletterId"`
	FeedID string `path:"feedId"`
	Body SubmittableFeedFields
}

type DeleteFeedInput struct {
	NewsletterID string `path:"newsletterId"`
	FeedID string `path:"feedId"`
}

type BaseFeed struct {
	Id string
	Name string
	URL string
	LastRetrievedAt time.Time
}

type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Description string `xml:"description"`
	Items []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Description string `xml:"description"`
	PubDate string `xml:"pubDate"`
}

type FeedView struct {
	Title string
	Items []RSSItem
}