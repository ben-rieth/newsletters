package email

import (
	"bytes"
	"context"
	"html/template"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/resend/resend-go/v3"
)

type EmailService interface {
	Send (ctx context.Context, subject, sender, recipient, body string) (*SendResult, error)
	AssembleEmail (templateName string, arguments map[string]any) (string, error)
}

type SendResult struct {
 	ID string
	Time time.Time
}

type ResendEmailService struct {
	resendClient *resend.Client
	tmpl *template.Template
}

func NewResendEmailService(globalConfig *config.Config, tmpl *template.Template) *ResendEmailService {
	resendClient := resend.NewClient(globalConfig.ResendAPIKey)
	return &ResendEmailService{resendClient, tmpl}
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

	sendTime := time.Now()
	return &SendResult{
		ID: sent.Id,
		Time: sendTime,
	}, err
}

func (s *ResendEmailService) AssembleEmail(templateName string, arguments map[string]any) (string, error) {
	buffer := new(bytes.Buffer)
	err := s.tmpl.ExecuteTemplate(buffer, templateName, arguments)

	if err != nil {
		return "", err
	}

	htmlString := buffer.String()
	return htmlString, nil
}