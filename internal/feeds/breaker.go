package feeds

import (
	"context"
	"errors"
	"fmt"
	"time"

	dbutil "github.com/ben-rieth/newsletter-api/internal/db"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/utils"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	failureWindow       = 6 * time.Hour
	failureThreshold    = 5
	baseDisableDuration = 24 * time.Hour
	maxDisableDuration  = 30 * 24 * time.Hour
	maxBackoffDoublings = 10
)

var ErrFeedDisabled = fmt.Errorf(
	"%w: feed is paused after repeated fetch failures", utils.UserError,
)

func disableDurationFor(priorDisables int32) time.Duration {
	doublings := min(max(priorDisables, 0), maxBackoffDoublings)

	duration := baseDisableDuration << doublings
	if duration > maxDisableDuration {
		return maxDisableDuration
	}

	return duration
}

func (s *FeedService) fetchFeed(ctx context.Context, feedId, url string) (*FetchFeedResult, error) {
	priorDisables, err := s.checkBreaker(ctx, feedId, url)
	if err != nil {
		return nil, err
	}

	result, err := s.rssService.FetchFeed(ctx, url)

	var fetchErr *FetchError
	if errors.As(err, &fetchErr) {
		s.recordFailure(ctx, feedId, url, fetchErr, priorDisables)
	}

	if err != nil {
		return nil, err
	}

	if feedId != "" {
		if resetErr := s.queries.ResetFeedBreaker(ctx, feedId); resetErr != nil {
			wideLog.AddErrorField(ctx, resetErr)
		}
	}

	return result, nil
}

func (s *FeedService) checkBreaker(ctx context.Context, feedId, url string) (int32, error) {
	if feedId == "" {
		count, err := s.queries.CountRecentUrlFailures(ctx, db.CountRecentUrlFailuresParams{
			Url:        url,
			OccurredAt: time.Now().Add(-failureWindow),
		})

		if err != nil {
			wideLog.AddErrorField(ctx, err)
			return 0, nil
		}

		if count >= failureThreshold {
			return 0, ErrFeedDisabled
		}

		return 0, nil
	}

	state, err := s.queries.GetFeedBreakerState(ctx, feedId)
	if err != nil {
		wideLog.AddErrorField(ctx, err)
		return 0, nil
	}

	if state.DisabledUntil.Valid && state.DisabledUntil.Time.After(time.Now()) {
		wideLog.AddLogField(ctx, "feedBreakerOpen", true)
		return state.DisableCount, ErrFeedDisabled
	}

	return state.DisableCount, nil
}

func (s *FeedService) recordFailure(
	ctx context.Context,
	feedId, url string,
	fetchErr *FetchError,
	priorDisables int32,
) {
	statusCode := pgtype.Int4{
		Int32: int32(fetchErr.StatusCode),
		Valid: fetchErr.StatusCode > 0,
	}

	if feedId == "" {
		err := s.queries.RecordUnknownFeedFetchFailure(ctx, db.RecordUnknownFeedFetchFailureParams{
			Url:        url,
			Kind:       fetchErr.Kind,
			StatusCode: statusCode,
			Message:    fetchErr.Message,
		})

		if err != nil {
			wideLog.AddErrorField(ctx, err)
		}

		return
	}

	feedUUID, err := dbutil.ToUUID(feedId)
	if err != nil {
		wideLog.AddErrorField(ctx, err)
		return
	}

	err = s.queries.RecordFeedFetchFailure(ctx, db.RecordFeedFetchFailureParams{
		FeedID:     feedUUID,
		Url:        url,
		Kind:       fetchErr.Kind,
		StatusCode: statusCode,
		Message:    fetchErr.Message,
	})

	if err != nil {
		wideLog.AddErrorField(ctx, err)
		return
	}

	if priorDisables == 0 {
		count, err := s.queries.CountRecentFeedFailures(ctx, db.CountRecentFeedFailuresParams{
			FeedID:     feedUUID,
			OccurredAt: time.Now().Add(-failureWindow),
		})

		if err != nil {
			wideLog.AddErrorField(ctx, err)
			return
		}

		wideLog.AddLogField(ctx, "feedRecentFailures", count)

		if count < failureThreshold {
			return
		}
	}

	disabledUntil := time.Now().Add(disableDurationFor(priorDisables))
	err = s.queries.DisableFeedUntil(ctx, db.DisableFeedUntilParams{
		ID:            feedId,
		DisabledUntil: pgtype.Timestamptz{Time: disabledUntil, Valid: true},
	})

	if err != nil {
		wideLog.AddErrorField(ctx, err)
		return
	}

	wideLog.AddLogField(ctx, "feedDisabledUntil", disabledUntil)
	wideLog.AddLogField(ctx, "feedDisableCount", priorDisables+1)
}
