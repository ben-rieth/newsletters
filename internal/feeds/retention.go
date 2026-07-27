package feeds

import (
	"context"
	"log"
	"log/slog"
	"time"

	dbutil "github.com/ben-rieth/newsletter-api/internal/db"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FailurePruner struct {
	queries *db.Queries
	db      *pgxpool.Pool
}

func NewFailurePruner(queries *db.Queries, db *pgxpool.Pool) *FailurePruner {
	return &FailurePruner{queries, db}
}

func (p *FailurePruner) KickOff(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// A daily ticker resets on every deploy, so a service shipping more than once
		// a day would never reach its first tick without this.
		p.pruneWithContext()

		for {
			select {
			case <-ticker.C:
				p.pruneWithContext()
			case <-ctx.Done():
				log.Println("Shutting down feed failure pruner")
				return
			}
		}
	}()
}

func (p *FailurePruner) pruneWithContext() {
	pruneCtx := context.Background()
	pruneCtx, wl := wideLog.CreateWideLogAndAddToContext(pruneCtx)

	wl.AddLogField("pruneId", uuid.New())

	startTime := time.Now()
	err := p.prune(pruneCtx)
	wl.AddLogField("duration", time.Since(startTime).String())

	level := slog.LevelInfo
	if err != nil {
		wl.AddErrorField(err)
		level = slog.LevelError
	}

	wl.Slog(pruneCtx, level)
}

func (p *FailurePruner) prune(ctx context.Context) error {
	if err := dbutil.WaitForDB(ctx, p.db); err != nil {
		return err
	}

	cutoff := time.Now().Add(-failureRetention)
	wideLog.AddLogField(ctx, "failureCutoff", cutoff)

	deleted, err := p.queries.DeleteFeedFetchFailuresBefore(ctx, cutoff)
	if err != nil {
		return err
	}

	wideLog.AddLogField(ctx, "deletedFailures", deleted)

	return nil
}
