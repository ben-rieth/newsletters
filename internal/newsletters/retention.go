package newsletters

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

type IssuePruner struct {
	queries *db.Queries
	db      *pgxpool.Pool
}

func NewIssuePruner(queries *db.Queries, db *pgxpool.Pool) *IssuePruner {
	return &IssuePruner{queries, db}
}

func (p *IssuePruner) KickOff(ctx context.Context) {
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
				log.Println("Shutting down issue pruner")
				return
			}
		}
	}()
}

func (p *IssuePruner) pruneWithContext() {
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

	wl.SlogAs(pruneCtx, level, "Issue Prune")
}

func (p *IssuePruner) prune(ctx context.Context) error {
	if err := dbutil.WaitForDB(ctx, p.db); err != nil {
		return err
	}

	// Each user's cutoff comes from their own retention setting, so the window is
	// resolved per row inside the query rather than once out here.
	deleted, err := p.queries.DeleteExpiredIssues(ctx)
	if err != nil {
		return err
	}

	wideLog.AddLogField(ctx, "deletedIssues", deleted)

	return nil
}
