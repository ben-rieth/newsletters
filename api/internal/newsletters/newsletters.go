package newsletters

import (
	"time"

	"github.com/ben-rieth/newsletter-api/internal/db"
)

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