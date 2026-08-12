package newsletters

import (
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
)

type ExportableNewsletter struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Frequency    string                 `json:"frequency"`
	SendDay      int                    `json:"sendDay"`
	SendHour     int                    `json:"sendHour"`
	SendMinute   int                    `json:"sendMinute"`
	SendTimezone string                 `json:"sendTimezone"`
	Feeds        []feeds.ExportableFeed `json:"feeds"`
}

type NewsletterExport struct {
	Version    int                  `json:"version"`
	ExportedAt time.Time            `json:"exportedAt"`
	Newsletter ExportableNewsletter `json:"newsletter"`
}

type NewslettersExport struct {
	Version     int                    `json:"version"`
	ExportedAt  time.Time              `json:"exportedAt"`
	Newsletters []ExportableNewsletter `json:"newsletters"`
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
