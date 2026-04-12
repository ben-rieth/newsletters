package newsletters

import (
	"context"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/resend/resend-go/v3"
)

type SendResult struct {
 	ID string
	Time time.Time
}

type ResendEmailService struct {
	resendClient *resend.Client
}

func NewResendEmailService(globalConfig *config.Config) *ResendEmailService {
	resendClient := resend.NewClient(globalConfig.ResendAPIKey)
	return &ResendEmailService{resendClient}
}

func (s *ResendEmailService) Send (ctx context.Context, subject, sender, recipient, body string) (*SendResult, error) {
	params := &resend.SendEmailRequest{
		To: []string{recipient},
		From: sender,
		Subject: subject,
		Html: body,
	}

	sent, err := s.resendClient.Emails.SendWithContext(ctx, params)
	if err != nil {
		return nil, err
	}

	return &SendResult{
		ID: sent.Id,
		Time: time.Now(),
	}, err
}