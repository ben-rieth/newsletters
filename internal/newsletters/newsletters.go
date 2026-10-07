package newsletters

import (
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
)

const ExportVersion = 2

type ExportableNewsletter struct {
	ID            string                 `json:"id" required:"false"`
	Name          string                 `json:"name" minLength:"1"`
	Frequency     string                 `json:"frequency" enum:"daily,weekly,monthly"`
	SendDay       int                    `json:"sendDay" minimum:"0" maximum:"31"`
	SendHour      int                    `json:"sendHour" minimum:"0" maximum:"23"`
	SendMinute    int                    `json:"sendMinute" minimum:"0" maximum:"59"`
	SendTimezone  string                 `json:"sendTimezone"`
	Status        string                 `json:"status" enum:"active,inactive" required:"false"`
	SendWhenEmpty bool                   `json:"sendWhenEmpty" required:"false"`
	Feeds         []feeds.ExportableFeed `json:"feeds" maxItems:"500"`
}

type NewslettersExport struct {
	Version     int                    `json:"version" enum:"1,2"`
	ExportedAt  time.Time              `json:"exportedAt" required:"false"`
	Newsletters []ExportableNewsletter `json:"newsletters" maxItems:"100"`
}

func DbNewsletterToExportable(nl db.Newsletter, nlFeeds []feeds.ExportableFeed) ExportableNewsletter {
	return ExportableNewsletter{
		ID:            nl.ID,
		Name:          nl.Name,
		Frequency:     string(nl.Frequency),
		SendDay:       int(nl.SendDay),
		SendHour:      int(nl.SendHour),
		SendMinute:    int(nl.SendMinute),
		SendTimezone:  nl.SendTimezone,
		Status:        string(nl.Status),
		SendWhenEmpty: nl.SendWhenEmpty,
		Feeds:         nlFeeds,
	}
}

type Issue struct {
	IssueID        string       `json:"issueId"`
	NewsletterID   string       `json:"newsletterId"`
	NewsletterName string       `json:"newsletterName"`
	SentAt         time.Time    `json:"sentAt"`
	State          db.ItemState `json:"state" enum:"read,unread"`
	ItemCount      int32        `json:"itemCount"`
	UnreadCount    int32        `json:"unreadCount"`
	PreviewTitles  []string     `json:"previewTitles"`
}

type DetailedIssue struct {
	IssueID        string       `json:"issueId"`
	NewsletterID   string       `json:"newsletterId"`
	NewsletterName string       `json:"newsletterName"`
	SentAt         time.Time    `json:"sentAt"`
	State          db.ItemState `json:"state" enum:"read,unread"`
	Feeds          []IssueFeed  `json:"feeds"`
}

type IssueItem struct {
	ItemID      string       `json:"itemId"`
	Title       string       `json:"title"`
	Token       string       `json:"token"`
	State       db.ItemState `json:"state" enum:"read,unread"`
	PublishDate time.Time    `json:"publishDate"`
}

type IssueFeed struct {
	Title   string      `json:"title"`
	HtmlURL string      `json:"webUrl"`
	Items   []IssueItem `json:"items"`
}

func DbNewsletterToNewsletterType(newsletter db.Newsletter) Newsletter {
	var lastSentAt *time.Time
	if newsletter.LastSentAt.Valid {
		lastSentAt = &newsletter.LastSentAt.Time
	} else {
		lastSentAt = nil
	}

	var oneOffSendTime *time.Time
	var regularSendTime *time.Time
	if newsletter.OriginalNextSendTime.Valid {
		oneOff := newsletter.NextSendTime
		oneOffSendTime = &oneOff
		regularSendTime = &newsletter.OriginalNextSendTime.Time
	}

	return Newsletter{
		ID:              newsletter.ID,
		Name:            newsletter.Name,
		Frequency:       string(newsletter.Frequency),
		SendDay:         int(newsletter.SendDay),
		SendHour:        int(newsletter.SendHour),
		SendMinute:      int(newsletter.SendMinute),
		SendTimezone:    newsletter.SendTimezone,
		LastSentAt:      lastSentAt,
		NextSendTime:    newsletter.NextSendTime,
		OneOffSendTime:  oneOffSendTime,
		RegularSendTime: regularSendTime,
		CreatedAt:       newsletter.CreatedAt,
		UpdatedAt:       newsletter.UpdatedAt,
		Status:          string(newsletter.Status),
		SendWhenEmpty:   newsletter.SendWhenEmpty,
	}
}
