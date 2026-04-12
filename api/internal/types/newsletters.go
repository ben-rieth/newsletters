package types

import "time"

type Newsletter struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Frequency string `json:"frequency"`
	NextSendTime time.Time `json:"nextSendTime"`
	SendDay int `json:"sendDay"`
	SendHour int `json:"sendHour"`
	SendMinute int `json:"sendMinute"`
	SendTimezone string `json:"sendTimezone"`
	LastSentAt *time.Time `json:"lastSentAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListNewslettersOutput struct {
	Body []Newsletter
}

type SubmittableNewsletterFields struct {
	Name string `json:"name" minLength:"1"`
	Frequency string `json:"frequency" enum:"daily,weekly,monthly"`
	SendDay *int `json:"sendDay,omitempty" minimum:"0" maximum:"31"`
	SendHour int `json:"sendHour" minimum:"0" maximum:"23"`
	SendMinute int `json:"sendMinute" minimum:"0" maximum:"59"`
	SendTimezone string `json:"sendTimezone"`
}

type CreateNewsletterInput struct {
	Body SubmittableNewsletterFields
}

type BaseNewsletterInput struct {
	ID string `path:"id"`
}

type GetNewsletterOutput struct {
    Body Newsletter
}

type UpdateNewsletterInput struct {
	BaseNewsletterInput
	Body SubmittableNewsletterFields
}

type DueNewsletters map[string][]BaseFeed

type SendableNewsletter struct {
	ID string
	Name string
	Frequency string
	SendDay int
	SendHour int
	SendMinute int
	SendTimezone string
	LastSendTime time.Time
	Email string
	UserID string
	Feeds []BaseFeed
}