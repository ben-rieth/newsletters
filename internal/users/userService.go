package users

import (
	"context"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
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

func (s *UserService) DeleteUser(ctx context.Context, userId string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

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

	err = qtx.DeleteAllNewslettersForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteAllRefreshTokensForUser(ctx, userId)
	if err != nil {
		return err
	}

	err = qtx.DeleteUser(ctx, userId)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *UserService) UpdatePassword(ctx context.Context, userId, newPassword string) error {
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

	if err := qtx.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       userId,
		Password: string(hash),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
