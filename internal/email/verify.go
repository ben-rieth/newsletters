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
	"golang.org/x/time/rate"
)

type EmailVerifyService struct {
	queries      *db.Queries
	db           *pgxpool.Pool
	config       config.Config
	emailService EmailService
	sendCooldown *auth.KeyedLimiter
	dailySends   *auth.KeyedLimiter
}

func NewEmailVerifyService(
	ctx context.Context,
	queries *db.Queries,
	pool *pgxpool.Pool,
	config config.Config,
	emailService EmailService,
) *EmailVerifyService {
	return &EmailVerifyService{
		queries:      queries,
		db:           pool,
		config:       config,
		emailService: emailService,
		sendCooldown: auth.NewKeyedLimiter(ctx, rate.Every(verificationSendCooldown), 1),
		dailySends: auth.NewKeyedLimiter(
			ctx,
			rate.Every(verificationSendWindow/maxVerificationSendsPerWindow),
			maxVerificationSendsPerWindow,
		),
	}
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

// Every send path is reachable without a session, so these bound how hard the
// app can be used to flood an inbox or burn the sender's reputation.
const verificationSendCooldown = time.Minute
const maxVerificationSendsPerWindow = 10
const verificationSendWindow = time.Hour * 24

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

	// Silent like the lockout, and checked before the old code is replaced so a
	// refused send leaves the code already in the inbox working.
	if !s.hasSendBudget(userEmail) {
		wideLog.AddLogField(ctx, "verificationSendThrottled", true)
		return nil
	}

	verificationToken, err := auth.MakeVerificationToken()
	if err != nil {
		return err
	}
	wideLog.AddLogField(ctx, "didMakeVerificationToken", true)

	hashedToken := hashVerificationToken(verificationToken)

	emailHtml, err := s.emailService.AssembleEmail("verify-email.html", map[string]any{
		"Code": verificationToken,
	})
	if err != nil {
		return err
	}

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

	// Spent only once the new code is saved, so a failed attempt costs the user
	// nothing. Concurrent sends can all pass the check above; an extra email is
	// cheaper than reserving the budget and refunding it on every error path.
	s.spendSendBudget(userEmail)

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

// Budgets are keyed by recipient rather than by account, because the inbox is
// what a flood harms, and re-registering an unverified address replaces its
// account row. The cooldown is checked first so a refused send costs nothing
// from the daily budget.
func (s *EmailVerifyService) hasSendBudget(recipient string) bool {
	return s.sendCooldown.WouldAllow(recipient) && s.dailySends.WouldAllow(recipient)
}

func (s *EmailVerifyService) spendSendBudget(recipient string) {
	s.sendCooldown.Allow(recipient)
	s.dailySends.Allow(recipient)
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
