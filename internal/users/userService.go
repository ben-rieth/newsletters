package users

import (
	"context"
	"errors"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/jackc/pgx/v5"
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
// inbox still decides who gets it — so signing up again takes it over. Otherwise
// anyone could permanently squat a whitelisted address by registering it first.
func (s *UserService) CreateUser(ctx context.Context, email, password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	existing, err := s.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.queries.CreateUser(ctx, db.CreateUserParams{
			Email:    email,
			Password: string(hash),
		})
	}
	if err != nil {
		return "", err
	}

	if existing.EmailVerifiedAt.Valid {
		return "", EmailInUseError
	}

	// Taken over in place rather than recreated: the code already in the inbox
	// keeps working, and the verify lockout stays on the row instead of being
	// cleared by signing up again.
	id, err := s.queries.ReplaceUnverifiedUserPassword(ctx, db.ReplaceUnverifiedUserPasswordParams{
		ID:       existing.ID,
		Password: string(hash),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", EmailInUseError
	}

	return id, err
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
