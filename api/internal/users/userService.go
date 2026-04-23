package users

import (
	"context"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/jackc/pgx/v5/pgxpool"
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
