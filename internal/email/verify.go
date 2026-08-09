package email

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
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

// An 8 digit code is only worth guessing in bulk, so a token dies after a
// handful of wrong answers and the user has to request a fresh one.
const maxVerificationAttempts = 5

func (s *EmailVerifyService) SendVerificationEmail(ctx context.Context, userID, userEmail string) error {
	return s.sendCode(ctx, userID, userEmail, db.TokenPurposeEmailVerify, "Verify your email")
}

func (s *EmailVerifyService) SendEmailUpdateVerification(ctx context.Context, userID, newEmail string) error {
	return s.sendCode(ctx, userID, newEmail, db.TokenPurposeEmailUpdate, "Confirm your new email")
}

func (s *EmailVerifyService) sendCode(
	ctx context.Context,
	userID string,
	userEmail string,
	purpose db.TokenPurpose,
	subject string,
) error {
	err := s.queries.DeleteExistingTokensWithPurpose(ctx, db.DeleteExistingTokensWithPurposeParams{
		UserID:  userID,
		Purpose: purpose,
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
		Purpose:   purpose,
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

	s.emailService.BackgroundSend(
		ctx,
		subject,
		s.config.NewsletterSenderEmail.Address,
		userEmail,
		emailHtml,
	)
	return nil
}

func (s *EmailVerifyService) VerifyUserEmail(ctx context.Context, userId, code string) error {
	err := s.tryToUseToken(ctx, userId, code, db.TokenPurposeEmailVerify)
	if err != nil {
		return err
	}

	if err = s.queries.MarkUserEmailAsVerified(ctx, userId); err != nil {
		return err
	}

	return nil
}

func (s *EmailVerifyService) VerifyUserEmailUpdate(ctx context.Context, userId, code string) error {
	err := s.tryToUseToken(ctx, userId, code, db.TokenPurposeEmailUpdate)
	if err != nil {
		return err
	}

	if err = s.queries.MarkUserEmailUpdateAsVerified(ctx, userId); err != nil {
		return err
	}

	return nil
}

func (s *EmailVerifyService) tryToUseToken(ctx context.Context, userId, code string, purpose db.TokenPurpose) error {
	token, err := s.queries.FindUnexpiredToken(ctx, db.FindUnexpiredTokenParams{
		UserID:               userId,
		Purpose:              purpose,
		ExpiresAtGreaterThan: time.Now(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InvalidTokenError
		}

		return err
	}

	if subtle.ConstantTimeCompare([]byte(token.Code), []byte(hashVerificationToken(code))) != 1 {
		attempts, err := s.queries.RecordFailedTokenAttempt(ctx, token.ID)
		if err != nil {
			return err
		}
		wideLog.AddLogField(ctx, "verificationAttempts", attempts)

		if attempts >= maxVerificationAttempts {
			if err := s.queries.DeleteVerificationToken(ctx, token.ID); err != nil {
				return err
			}
		}

		return InvalidTokenError
	}

	if err = s.queries.DeleteVerificationToken(ctx, token.ID); err != nil {
		return err
	}

	return nil
}

func hashVerificationToken(token string) string {
	hashedToken := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashedToken[:])
}
