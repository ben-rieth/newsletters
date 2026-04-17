package newsletters

import (
	"time"

	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
)

type ExportableNewsletter struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Frequency string `json:"frequency"`
	SendDay int `json:"sendDay"`
	SendHour int `json:"sendHour"`
	SendMinute int `json:"sendMinute"`
	SendTimezone string `json:"sendTimezone"`
	Feeds []feeds.ExportableFeed `json:"feeds"`
}


type NewsletterExport struct {
	Version int `json:"version"`
	ExportedAt time.Time `json:"exportedAt"`
	Newsletter ExportableNewsletter `json:"newsletter"`
}

type NewslettersExport struct {
	Version int `json:"version"`
	ExportedAt time.Time `json:"exportedAt"`
	Newsletters []ExportableNewsletter `json:"newsletters"`
}

func DbNewsletterToNewsletterType (newsletter db.Newsletter) Newsletter {
	var lastSentAt *time.Time
	if newsletter.LastSentAt.Valid {
		lastSentAt = &newsletter.LastSentAt.Time
	} else {
		lastSentAt = nil
	}
	
	return Newsletter{
		ID: newsletter.ID,
		Name: newsletter.Name,
		Frequency: string(newsletter.Frequency),
		SendDay: int(newsletter.SendDay),
		SendHour: int(newsletter.SendHour),
		SendMinute: int(newsletter.SendMinute),
		SendTimezone: newsletter.SendTimezone,
		LastSentAt: lastSentAt,
		NextSendTime: newsletter.NextSendTime,
		CreatedAt: newsletter.CreatedAt,
		UpdatedAt: newsletter.UpdatedAt,
	}
}