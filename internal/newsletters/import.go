package newsletters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	dbgen "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/jobs"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxImportedFeeds          = 500
	concurrentFeedResolutions = 4
)

var ErrInvalidImport = errors.New("Invalid import")

type ImportResult struct {
	NewsletterIDs []string `json:"newsletterIds"`
	PendingFeeds  int      `json:"pendingFeeds"`
}

type ImportService struct {
	queries     *dbgen.Queries
	db          *pgxpool.Pool
	feedService *feeds.FeedService
	jobQueue    *jobs.JobQueue
}

func NewImportService(
	queries *dbgen.Queries,
	db *pgxpool.Pool,
	feedService *feeds.FeedService,
	jobQueue *jobs.JobQueue,
) *ImportService {
	return &ImportService{queries, db, feedService, jobQueue}
}

func (s *ImportService) Import(ctx context.Context, userID string, export NewslettersExport) (*ImportResult, error) {
	if err := validateExport(export); err != nil {
		return nil, err
	}

	knownFeeds, err := s.knownFeedIDs(ctx, export)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	result := &ImportResult{NewsletterIDs: []string{}}

	for _, nl := range export.Newsletters {
		nextSendTime, err := ComputeNextSendTime(
			dbgen.Frequency(nl.Frequency),
			nl.SendDay, nl.SendHour, nl.SendMinute,
			nl.SendTimezone, time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("%w: %q has an invalid timezone", ErrInvalidImport, nl.Name)
		}

		newsletterID, err := qtx.ImportNewsletter(ctx, dbgen.ImportNewsletterParams{
			Name:          nl.Name,
			Frequency:     dbgen.Frequency(nl.Frequency),
			SendDay:       int32(nl.SendDay),
			SendHour:      int32(nl.SendHour),
			SendMinute:    int32(nl.SendMinute),
			SendTimezone:  nl.SendTimezone,
			NextSendTime:  nextSendTime,
			Status:        statusOrActive(nl.Status),
			SendWhenEmpty: nl.SendWhenEmpty,
			UserID:        userID,
		})
		if err != nil {
			return nil, err
		}

		pending, err := importFeeds(ctx, qtx, newsletterID, userID, nl.Feeds, knownFeeds)
		if err != nil {
			return nil, err
		}

		result.NewsletterIDs = append(result.NewsletterIDs, newsletterID)
		result.PendingFeeds += pending
	}

	pendingIDs, err := qtx.GetPendingFeedImportIdsForNewsletters(ctx, result.NewsletterIDs)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.enqueueResolution(pendingIDs)

	return result, nil
}

func validateExport(export NewslettersExport) error {
	total := 0
	for _, nl := range export.Newsletters {
		total += len(nl.Feeds)
	}

	if total > maxImportedFeeds {
		return fmt.Errorf("%w: at most %d feeds can be imported at once", ErrInvalidImport, maxImportedFeeds)
	}

	for _, nl := range export.Newsletters {
		if err := validateSchedule(nl); err != nil {
			return err
		}
	}

	return nil
}

func validateSchedule(nl ExportableNewsletter) error {
	if nl.SendTimezone == "" {
		return fmt.Errorf("%w: %q has no timezone", ErrInvalidImport, nl.Name)
	}

	if err := ValidateSendDay(dbgen.Frequency(nl.Frequency), &nl.SendDay); err != nil {
		return fmt.Errorf("%w: %q: %w", ErrInvalidImport, nl.Name, err)
	}

	return nil
}

func statusOrActive(status string) dbgen.NewsletterStatus {
	if status == string(dbgen.NewsletterStatusInactive) {
		return dbgen.NewsletterStatusInactive
	}
	return dbgen.NewsletterStatusActive
}

func (s *ImportService) knownFeedIDs(ctx context.Context, export NewslettersExport) (map[string]string, error) {
	var urls []string
	for _, nl := range export.Newsletters {
		for _, feed := range nl.Feeds {
			urls = append(urls, feed.URL)
		}
	}

	rows, err := s.queries.GetFeedIdsForUrls(ctx, urls)
	if err != nil {
		return nil, err
	}

	known := make(map[string]string, len(rows))
	for _, row := range rows {
		known[row.Url] = row.FeedID
	}
	return known, nil
}

// Feeds already in the database are attached straight away; only unknown URLs
// need fetching, which is slow enough to belong in the background.
func importFeeds(
	ctx context.Context,
	qtx *dbgen.Queries,
	newsletterID, userID string,
	exported []feeds.ExportableFeed,
	knownFeeds map[string]string,
) (int, error) {
	var unknown []feeds.ExportableFeed
	seenURLs := make(map[string]bool, len(exported))
	for _, feed := range exported {
		if seenURLs[feed.URL] {
			continue
		}
		seenURLs[feed.URL] = true

		feedID, ok := knownFeeds[feed.URL]
		if !ok {
			unknown = append(unknown, feed)
			continue
		}

		err := attachFeed(ctx, qtx, dbgen.AddImportedNewsletterFeedParams{
			NewsletterID: newsletterID,
			FeedID:       feedID,
			UserID:       userID,
			Alias:        feed.Alias,
			Status:       statusOrActive(feed.Status),
		}, feed.Filters)
		if err != nil {
			return 0, err
		}
	}

	feedImports, err := buildFeedImports(newsletterID, userID, unknown)
	if err != nil {
		return 0, err
	}

	if _, err := qtx.CreateFeedImports(ctx, feedImports); err != nil {
		return 0, err
	}

	return len(feedImports), nil
}

func attachFeed(
	ctx context.Context,
	qtx *dbgen.Queries,
	feed dbgen.AddImportedNewsletterFeedParams,
	filters []feeds.ExportableFilter,
) error {
	newsletterFeedID, err := qtx.AddImportedNewsletterFeed(ctx, feed)
	alreadyInNewsletter := errors.Is(err, pgx.ErrNoRows)
	if alreadyInNewsletter {
		return nil
	}
	if err != nil {
		return err
	}

	for _, f := range filters {
		err := qtx.AddFeedFilter(ctx, dbgen.AddFeedFilterParams{
			NewsletterFeedID: newsletterFeedID,
			UserID:           feed.UserID,
			Field:            f.Field,
			Operator:         f.Operator,
			Pattern:          f.Pattern,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func buildFeedImports(newsletterID, userID string, exported []feeds.ExportableFeed) ([]dbgen.CreateFeedImportsParams, error) {
	params := make([]dbgen.CreateFeedImportsParams, 0, len(exported))
	for _, feed := range exported {
		filters := feed.Filters
		if filters == nil {
			filters = []feeds.ExportableFilter{}
		}

		encodedFilters, err := json.Marshal(filters)
		if err != nil {
			return nil, err
		}

		params = append(params, dbgen.CreateFeedImportsParams{
			NewsletterID: newsletterID,
			UserID:       userID,
			Url:          feed.URL,
			Alias:        feed.Alias,
			Status:       statusOrActive(feed.Status),
			Filters:      encodedFilters,
		})
	}
	return params, nil
}

func (s *ImportService) RetryFeedImport(ctx context.Context, importID, newsletterID, userID string) (bool, error) {
	updated, err := s.queries.RetryFeedImport(ctx, dbgen.RetryFeedImportParams{
		ID:           importID,
		NewsletterID: newsletterID,
		UserID:       userID,
	})
	if err != nil || updated == 0 {
		return false, err
	}

	s.enqueueResolution([]string{importID})
	return true, nil
}

// Imports interrupted by a restart are still pending in the database but no
// longer queued, since the job queue only lives in memory.
func (s *ImportService) ResumePendingImports(ctx context.Context) error {
	ids, err := s.queries.GetPendingFeedImportIds(ctx)
	if err != nil {
		return err
	}

	s.enqueueResolution(ids)
	return nil
}

// A single job per batch rather than one per feed: the queue's buffer is small
// and Enqueue blocks when it's full, which would stall the import request.
func (s *ImportService) enqueueResolution(importIDs []string) {
	if len(importIDs) == 0 {
		return
	}

	s.jobQueue.Enqueue(func(ctx context.Context) {
		s.resolveFeedImports(ctx, importIDs)
	})
}

func (s *ImportService) resolveFeedImports(ctx context.Context, importIDs []string) {
	wideLog.AddLogField(ctx, "job", "resolve-feed-imports")
	wideLog.AddLogField(ctx, "feedImportCount", len(importIDs))

	sem := make(chan struct{}, concurrentFeedResolutions)
	var wg sync.WaitGroup

	for _, id := range importIDs {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()

			if err := s.resolveFeedImport(ctx, id); err != nil {
				wideLog.AddErrorField(ctx, err)
			}
		})
	}

	wg.Wait()
}

func (s *ImportService) resolveFeedImport(ctx context.Context, importID string) error {
	imp, err := s.queries.GetPendingFeedImport(ctx, importID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	feed, err := s.feedService.GetFeedMetaData(ctx, imp.Url, true)
	if err != nil && !isFetchFailure(err) {
		// Concurrent resolution of a duplicate URL races the feed insert; the retry hits the cache.
		feed, err = s.feedService.GetFeedMetaData(ctx, imp.Url, true)
	}
	if err != nil {
		wideLog.AddArrayField(ctx, "feedImportFailures", imp.Url)
		return s.queries.MarkFeedImportFailed(ctx, dbgen.MarkFeedImportFailedParams{
			ID:    importID,
			Error: feeds.DescribeFetchError(err),
		})
	}

	if err := s.addImportedFeed(ctx, imp, feed.Id); err != nil {
		markErr := s.queries.MarkFeedImportFailed(ctx, dbgen.MarkFeedImportFailedParams{
			ID:    importID,
			Error: "Could not be added",
		})
		return errors.Join(err, markErr)
	}

	return nil
}

func isFetchFailure(err error) bool {
	var fetchErr *feeds.FetchError
	return errors.As(err, &fetchErr) || errors.Is(err, feeds.ErrFeedDisabled)
}

func (s *ImportService) addImportedFeed(ctx context.Context, imp dbgen.NewsletterFeedImport, feedID string) error {
	var filters []feeds.ExportableFilter
	if err := json.Unmarshal(imp.Filters, &filters); err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// Claiming the row first means a Remove that lands while the feed was being
	// fetched wins, instead of the feed reappearing afterwards.
	claimed, err := qtx.DeleteResolvedFeedImport(ctx, imp.ID)
	if err != nil || claimed == 0 {
		return err
	}

	err = attachFeed(ctx, qtx, dbgen.AddImportedNewsletterFeedParams{
		NewsletterID: imp.NewsletterID,
		FeedID:       feedID,
		UserID:       imp.UserID,
		Alias:        imp.Alias,
		Status:       imp.Status,
	}, filters)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
