package users

import (
	"context"
	"errors"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	queries *db.Queries
	db      *pgxpool.Pool
}

func NewUserService(queries *db.Queries, db *pgxpool.Pool) *UserService {
	return &UserService{queries, db}
}

var EmailInUseError = errors.New("Email already in use")

func (s *UserService) DeleteUser(ctx context.Context, userId string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err := deleteUser(ctx, s.queries.WithTx(tx), userId); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Every table referencing app_user is ON DELETE RESTRICT, so the user row can
// only go once all of these have. Order matters.
func deleteUser(ctx context.Context, qtx *db.Queries, userId string) error {
	err := qtx.DeleteFeedImportsForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteNewsletterFeedItemStatuses(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteFeedFiltersForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteAllNewsletterFeedsForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteAllIssueItemsForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteAllIssuesForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteAllNewslettersForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteAllRefreshTokensForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteAllVerificationTokensForUser(ctx, userId)
	if err != nil {
		return err
	}

	return qtx.DeleteUser(ctx, userId)
}

// CreateUser returns EmailInUseError only for a verified account. An account
// that was registered but never verified is nobody's yet — whoever holds the
// inbox still decides who gets it — so signing up again replaces it. Otherwise
// anyone could permanently squat a whitelisted address by registering it first.
func (s *UserService) CreateUser(ctx context.Context, email, password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	existing, err := qtx.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	var carriedAttempts int32
	var carriedLock pgtype.Timestamptz

	if err == nil {
		if existing.EmailVerifiedAt.Valid {
			return "", EmailInUseError
		}

		// The replacement row would start with a clean verification budget, which
		// would make signing up again the way to clear a lockout.
		carriedAttempts, carriedLock = existing.VerifyAttempts, existing.VerifyLockedUntil

		if err := deleteUser(ctx, qtx, existing.ID); err != nil {
			return "", err
		}
	}

	id, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:    email,
		Password: string(hash),
	})
	if err != nil {
		return "", err
	}

	if carriedAttempts != 0 || carriedLock.Valid {
		if err := qtx.CarryVerifyLockout(ctx, db.CarryVerifyLockoutParams{
			ID:                id,
			VerifyAttempts:    carriedAttempts,
			VerifyLockedUntil: carriedLock,
		}); err != nil {
			return "", err
		}
	}

	return id, tx.Commit(ctx)
}

func (s *UserService) UpdatePassword(ctx context.Context, userId, newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteAllRefreshTokensForUser(ctx, userId); err != nil {
		return err
	}

	// Refresh tokens are rows and are gone above. Access tokens are not, so the
	// old password's sessions only die if the cutoff moves too.
	if err := qtx.InvalidateUserSessions(ctx, userId); err != nil {
		return err
	}

	if err := qtx.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       userId,
		Password: string(hash),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
