package types

import (
	"context"
)

type EmailService interface {
	Send (ctx context.Context, subject, sender, recipient, body string) (*SendResult, error)
}