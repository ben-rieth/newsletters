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
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailVerifyService struct {
	queries      *db.Queries
	db           *pgxpool.Pool
	config       config.Config
	emailService EmailService
}

func NewEmailVerifyService(queries *db.Queries, pool *pgxpool.Pool, config config.Config, emailService EmailService) *EmailVerifyService {
	return &EmailVerifyService{queries, pool, config, emailService}
}

var InvalidTokenError = errors.New("Invalid token")

var EmailUpdateNoLongerValidError = errors.New("Email update is no longer valid")

// An 8 digit code is only worth guessing in bulk, so a token dies after a
// handful of wrong answers and the user has to request a fresh one.
const maxVerificationAttempts = 5

// Asking for a fresh code resets the per-token budget above, so the budget that
// actually bounds guessing has to live on the account, where neither a resend nor
// a new IP address can shake it off.
const maxAccountVerifyAttempts = 10
const verifyLockout = time.Minute * 15

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
	locked, err := s.isVerifyLocked(ctx, userID)
	if err != nil {
		return err
	}

	// Handing out a fresh code while locked out would be the way around the
	// lockout. Silent, so the caller cannot tell a lockout from a normal send.
	if locked {
		wideLog.AddLogField(ctx, "verifyLockedOut", true)
		return nil
	}

	verificationToken, err := auth.MakeVerificationToken()
	if err != nil {
		return err
	}
	wideLog.AddLogField(ctx, "didMakeVerificationToken", true)

	hashedToken := hashVerificationToken(verificationToken)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// Replacing the old code has to be atomic, or a race leaves two live codes for
	// one purpose and which of them works becomes arbitrary.
	err = qtx.DeleteExistingTokensWithPurpose(ctx, db.DeleteExistingTokensWithPurposeParams{
		UserID:  userID,
		Purpose: purpose,
	})
	if err != nil {
		return err
	}

	err = qtx.SaveVerificationToken(ctx, db.SaveVerificationTokenParams{
		UserID:    userID,
		Code:      hashedToken,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(time.Minute * 10),
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
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

	if _, err = s.queries.MarkUserEmailUpdateAsVerified(ctx, userId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EmailUpdateNoLongerValidError
		}

		return err
	}

	return nil
}

func (s *EmailVerifyService) isVerifyLocked(ctx context.Context, userID string) (bool, error) {
	lockedUntil, err := s.queries.GetUserVerifyLock(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return lockedUntil.Valid && lockedUntil.Time.After(time.Now()), nil
}

func (s *EmailVerifyService) tryToUseToken(ctx context.Context, userId, code string, purpose db.TokenPurpose) error {
	locked, err := s.isVerifyLocked(ctx, userId)
	if err != nil {
		return err
	}

	// Reported as an invalid token rather than as a lockout, so probing this
	// endpoint says nothing about the account behind it.
	if locked {
		wideLog.AddLogField(ctx, "verifyLockedOut", true)
		return InvalidTokenError
	}

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
		attempt, err := s.queries.RecordFailedVerifyAttempt(ctx, db.RecordFailedVerifyAttemptParams{
			MaxAttempts: maxAccountVerifyAttempts,
			LockedUntil: time.Now().Add(verifyLockout),
			ID:          userId,
		})
		if err != nil {
			return err
		}
		wideLog.AddLogField(ctx, "accountVerifyAttempts", attempt.VerifyAttempts)

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

	return s.queries.ResetVerifyAttempts(ctx, userId)
}

func hashVerificationToken(token string) string {
	hashedToken := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashedToken[:])
}
