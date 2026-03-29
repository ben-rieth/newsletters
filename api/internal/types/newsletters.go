package types

import "time"

type Newsletter struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Frequency string `json:"frequency"`
	SendDay *int `json:"send_day,omitempty"`
	SendHour int `json:"send_hour"`
	SendMinute int `json:"send_minute"`
	LastSentAt *time.Time `json:"last_sent_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListNewslettersOutput struct {
	Body []Newsletter
}

type CreateNewsletterInput struct {
	Body struct {
		Name string `json:"name" minLength:"1"`
		Frequency string `json:"frequency" enum:"daily,weekly,monthly"`
		SendDay *int `json:"send_day,omitempty" minimum:"0" maximum:"31"`
		SendHour int `json:"send_hour" minimum:"0" maximum:"23"`
		SendMinute int `json:"send_minute" minimum:"0" maximum:"59"`
	}
}

type CreateNewsletterOutput struct {
    Body Newsletter
}

type GetNewsletterInput struct {
    ID string `path:"id"`
}

type GetNewsletterOutput struct {
    Body Newsletter
}