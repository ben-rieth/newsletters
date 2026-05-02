package email

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/jackc/pgx/v5"
)

type EmailVerifyService struct {
	queries      *db.Queries
	config       config.Config
	emailService EmailService
}

func NewEmailVerifyService(queries *db.Queries, config config.Config, emailService EmailService) *EmailVerifyService {
	return &EmailVerifyService{queries, config, emailService}
}

var InvalidTokenError = errors.New("Invalid token")

func (s *EmailVerifyService) SendVerificationEmail(
	ctx context.Context,
	userID string,
	userEmail string,
) error {
	err := s.queries.DeleteExistingTokensWithPurpose(ctx, db.DeleteExistingTokensWithPurposeParams{
		UserID:  userID,
		Purpose: db.TokenPurposeEmailVerify,
	})
	if err != nil {
		return err
	}

	verificationToken, err := auth.MakeVerificationToken()
	if err != nil {
		return err
	}
	wideLog.AddLogField(ctx, "didMakeVerificationToken", true)

	hashedToken := hashVerificationToken(verificationToken)

	err = s.queries.SaveVerificationToken(ctx, db.SaveVerificationTokenParams{
		UserID:    userID,
		Code:      hashedToken,
		Purpose:   db.TokenPurposeEmailVerify,
		ExpiresAt: time.Now().Add(time.Minute * 10),
	})
	if err != nil {
		return err
	}

	emailHtml, err := s.emailService.AssembleEmail("verify-email.html", map[string]any{
		"Code": verificationToken,
	})
	if err != nil {
		return err
	}

	result, err := s.emailService.Send(
		ctx,
		"Verify your email",
		s.config.NewsletterSenderEmail.Address,
		userEmail,
		emailHtml,
	)
	if err != nil {
		return err
	}

	wideLog.AddLogField(ctx, "emailSendResult", result)
	return nil
}

func (s *EmailVerifyService) VerifyUserEmail(ctx context.Context, userId, code string) error {
	err := s.tryToUseToken(ctx, userId, code)
	if err != nil {
		return err
	}

	if err = s.queries.MarkUserEmailAsVerified(ctx, userId); err != nil {
		return err
	}

	return nil
}

func (s *EmailVerifyService) VerifyUserEmailUpdate(ctx context.Context, userId, code string) error {
	err := s.tryToUseToken(ctx, userId, code)
	if err != nil {
		return err
	}

	if err = s.queries.MarkUserEmailUpdateAsVerified(ctx, userId); err != nil {
		return err
	}

	return nil
}

func (s *EmailVerifyService) tryToUseToken(ctx context.Context, userId, code string) error {
	hashedToken := hashVerificationToken(code)
	_, err := s.queries.FindValidToken(ctx, db.FindValidTokenParams{
		UserID:               userId,
		Code:                 hashedToken,
		Purpose:              db.TokenPurposeEmailVerify,
		ExpiresAtGreaterThan: time.Now(),
	})
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return InvalidTokenError
		}

		return err
	}

	if err = s.queries.DeleteExistingTokensWithPurpose(ctx, db.DeleteExistingTokensWithPurposeParams{
		UserID:  userId,
		Purpose: db.TokenPurposeEmailVerify,
	}); err != nil {
		return err
	}

	return nil
}

func hashVerificationToken(token string) string {
	hashedToken := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashedToken[:])
}
