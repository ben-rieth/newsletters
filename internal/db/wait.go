package db

import (
	"context"
	"fmt"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/jackc/pgx/v5/pgxpool"
)

func WaitForDB(ctx context.Context, db *pgxpool.Pool) error {
	maxAttempts := 3
	delay := 1 * time.Second

	for i := range maxAttempts {
		err := db.Ping(ctx)
		if err == nil {
			return nil
		}

		wideLog.AddArrayField(ctx, "waitForDbAttempts", fmt.Errorf("DB not ready on ping attempt %d: %w", i, err))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			delay = delay * 2
		}
	}

	return fmt.Errorf("Could not reach database after %d attempts", maxAttempts)
}
