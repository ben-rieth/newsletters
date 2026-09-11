package email

import (
	"bytes"
	"context"
	"html/template"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/jobs"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/resend/resend-go/v3"
)

type EmailService interface {
	Send(ctx context.Context, subject, sender, recipient, body string) (*SendResult, error)
	SendIdempotent(ctx context.Context, subject, sender, recipient, body, idempotencyKey string) (*SendResult, error)
	BackgroundSend(ctx context.Context, subject, sender, recipient, body string)
	AssembleEmail(templateName string, arguments map[string]any) (string, error)
}

type SendResult struct {
	ID   string
	Time time.Time
}

type ResendEmailService struct {
	resendClient *resend.Client
	tmpl         *template.Template
	jobQueue     *jobs.JobQueue
}

func NewResendEmailService(globalConfig *config.Config, tmpl *template.Template, jobQueue *jobs.JobQueue) *ResendEmailService {
	resendClient := resend.NewClient(globalConfig.ResendAPIKey)
	return &ResendEmailService{resendClient, tmpl, jobQueue}
}

func (s *ResendEmailService) Send(ctx context.Context, subject, sender, recipient, body string) (*SendResult, error) {
	return s.SendIdempotent(ctx, subject, sender, recipient, body, "")
}

// Resend only dedupes when the key is non-empty, so an empty key sends normally.
func (s *ResendEmailService) SendIdempotent(
	ctx context.Context,
	subject, sender, recipient, body, idempotencyKey string,
) (*SendResult, error) {
	params := &resend.SendEmailRequest{
		To:      []string{recipient},
		From:    sender,
		Subject: subject,
		Html:    body,
	}

	sent, err := s.resendClient.Emails.SendWithOptions(ctx, params, &resend.SendEmailOptions{
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, err
	}

	sendTime := time.Now()
	return &SendResult{
		ID:   sent.Id,
		Time: sendTime,
	}, nil
}

func (s *ResendEmailService) BackgroundSend(ctx context.Context, subject, sender, recipient, body string) {
	s.jobQueue.Enqueue(func(ctx context.Context) {
		result, err := s.Send(ctx, subject, sender, recipient, body)
		if err != nil {
			wideLog.AddErrorField(ctx, err)
		}

		wideLog.AddLogField(ctx, "sendResult", result)
	})
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
