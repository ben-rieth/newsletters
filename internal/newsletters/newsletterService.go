package newsletters

import (
	"context"
	"errors"
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db"
	dbgen "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Newsletter struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Frequency       string     `json:"frequency"`
	NextSendTime    time.Time  `json:"nextSendTime"`
	OneOffSendTime  *time.Time `json:"oneOffSendTime,omitempty"`
	RegularSendTime *time.Time `json:"regularSendTime,omitempty"`
	SendDay         int        `json:"sendDay"`
	SendHour        int        `json:"sendHour"`
	SendMinute      int        `json:"sendMinute"`
	SendTimezone    string     `json:"sendTimezone"`
	LastSentAt      *time.Time `json:"lastSentAt,omitempty"`
	Status          string     `json:"status"`
	SendWhenEmpty   bool       `json:"sendWhenEmpty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

var ErrNewsletterNotFound = errors.New("Newsletter not found")

type NewsletterService struct {
	queries *dbgen.Queries
	db      *pgxpool.Pool
}

func NewNewsletterService(queries *dbgen.Queries, db *pgxpool.Pool) *NewsletterService {
	return &NewsletterService{queries, db}
}

func (s *NewsletterService) GetSendableNewsletter(ctx context.Context, newsletterID string) (*SendableNewsletter, error) {
	newsletterResult, err := s.queries.GetSendableNewsletter(ctx, newsletterID)
	if err != nil {
		return nil, err
	}

	var lastSentAt time.Time
	if newsletterResult.LastSentAt.Valid {
		lastSentAt = newsletterResult.LastSentAt.Time
	} else {
		computedLastSentAt, lastErr := ComputeLastSendTime(
			newsletterResult.Frequency,
			int(newsletterResult.SendDay), int(newsletterResult.SendHour), int(newsletterResult.SendMinute),
			newsletterResult.SendTimezone, time.Now(),
		)

		if lastErr != nil {
			wideLog.AddErrorField(ctx, lastErr)
			return nil, lastErr
		}

		lastSentAt = computedLastSentAt
	}

	feedsResult, err := s.queries.GetSendableFeedsForNewsletter(ctx, newsletterID)
	if err != nil {
		return nil, err
	}

	var sendableFeeds []feeds.BaseFeed
	for _, row := range feedsResult {
		baseFeed := feeds.BaseFeed{
			GlobalFeedId:     row.GlobalFeedID,
			NewsletterFeedId: row.NewsletterFeedID,
			Name:             row.Title,
			URL:              row.Url,
			HtmlURL:          row.HtmlUrl,
			LastRetrievedAt:  row.LastRetrievedAt,
		}

		if len(row.Alias) > 0 {
			baseFeed.Name = row.Alias
		}

		sendableFeeds = append(
			sendableFeeds,
			baseFeed,
		)
	}

	return &SendableNewsletter{
		ID:               newsletterResult.ID,
		Name:             newsletterResult.Name,
		Frequency:        string(newsletterResult.Frequency),
		SendDay:          int(newsletterResult.SendDay),
		SendHour:         int(newsletterResult.SendHour),
		SendMinute:       int(newsletterResult.SendMinute),
		SendTimezone:     newsletterResult.SendTimezone,
		Email:            newsletterResult.Email,
		UserID:           newsletterResult.UserID,
		LastSendTime:     lastSentAt,
		NextSendTime:     newsletterResult.NextSendTime,
		UnsubscribeToken: newsletterResult.UnsubscribeToken,
		SendWhenEmpty:    newsletterResult.SendWhenEmpty,
		IsOneOffSend:     newsletterResult.IsOneOffSend,
		Feeds:            sendableFeeds,
	}, nil
}

func nextSendTimeFor(nl *SendableNewsletter) (time.Time, error) {
	return ComputeNextSendTime(
		dbgen.Frequency(nl.Frequency), int(nl.SendDay), int(nl.SendHour), int(nl.SendMinute), nl.SendTimezone,
		time.Now(),
	)
}

func (s *NewsletterService) UpdateSendTimes(
	ctx context.Context,
	nl *SendableNewsletter,
	sentAt time.Time,
) error {
	nextSendTime, err := nextSendTimeFor(nl)
	if err != nil {
		return err
	}

	return s.queries.UpdateNewsletterSendTimes(ctx, dbgen.UpdateNewsletterSendTimesParams{
		ID:           nl.ID,
		UserID:       nl.UserID,
		NextSendTime: nextSendTime,
		LastSentAt:   db.ToTimestamp(&sentAt),
	})
}

// last_sent_at only anchors on the first skip, so the skipped window is still covered by the next send.
func (s *NewsletterService) SkipSend(ctx context.Context, nl *SendableNewsletter) error {
	nextSendTime, err := nextSendTimeFor(nl)
	if err != nil {
		return err
	}

	lastSendTime := nl.LastSendTime
	return s.queries.SkipNewsletterSend(ctx, dbgen.SkipNewsletterSendParams{
		ID:           nl.ID,
		UserID:       nl.UserID,
		NextSendTime: nextSendTime,
		LastSentAt:   db.ToTimestamp(&lastSendTime),
	})
}

func (s *NewsletterService) DeleteNewsletter(
	ctx context.Context,
	id, userId string,
) error {
	return s.DeleteNewsletters(ctx, []string{id}, userId)
}

func (s *NewsletterService) DeleteNewsletters(
	ctx context.Context,
	ids []string,
	userId string,
) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	for _, id := range ids {
		if err := deleteNewsletterWith(ctx, qtx, id, userId); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *NewsletterService) UpdateStatuses(
	ctx context.Context,
	ids []string,
	userId string,
	status dbgen.NewsletterStatus,
) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	previous, err := qtx.ListNewslettersByIds(ctx, dbgen.ListNewslettersByIdsParams{
		UserID: userId,
		Ids:    ids,
	})
	if err != nil {
		return err
	}

	updated, err := qtx.UpdateNewslettersStatus(ctx, dbgen.UpdateNewslettersStatusParams{
		Status: status,
		UserID: userId,
		Ids:    ids,
	})
	if err != nil {
		return err
	}

	if updated != int64(len(ids)) {
		return ErrNewsletterNotFound
	}

	if status == dbgen.NewsletterStatusActive {
		var resumed []dbgen.Newsletter
		for _, nl := range previous {
			if nl.Status == dbgen.NewsletterStatusInactive {
				resumed = append(resumed, nl)
			}
		}

		if err := skipMissedSends(ctx, qtx, resumed, userId); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// A paused newsletter's next_send_time goes stale; left alone, resuming would
// send immediately instead of waiting for the next scheduled slot.
func skipMissedSends(ctx context.Context, qtx *dbgen.Queries, nls []dbgen.Newsletter, userId string) error {
	now := time.Now()
	for _, nl := range nls {
		if !nl.NextSendTime.Before(now) {
			continue
		}

		next, err := ComputeNextSendTime(
			nl.Frequency, int(nl.SendDay), int(nl.SendHour), int(nl.SendMinute), nl.SendTimezone,
			now,
		)
		if err != nil {
			return err
		}

		err = qtx.RescheduleNewsletter(ctx, dbgen.RescheduleNewsletterParams{
			NextSendTime: next,
			ID:           nl.ID,
			UserID:       userId,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteNewsletterWith(ctx context.Context, qtx *dbgen.Queries, id, userId string) error {
	err := qtx.DeleteFeedImportsForNewsletter(ctx, dbgen.DeleteFeedImportsForNewsletterParams{
		NewsletterID: id,
		UserID:       userId,
	})
	if err != nil {
		return err
	}

	// Everything pointing at the newsletter is ON DELETE RESTRICT, so the filters
	// and the sent archive have to be cleared before the rows they hang off of.
	err = qtx.DeleteFeedFiltersForNewsletter(ctx, dbgen.DeleteFeedFiltersForNewsletterParams{
		NewsletterID: id,
		UserID:       userId,
	})
	if err != nil {
		return err
	}

	err = qtx.DeleteIssuesForNewsletter(ctx, dbgen.DeleteIssuesForNewsletterParams{
		NewsletterID: id,
		UserID:       userId,
	})
	if err != nil {
		return err
	}

	err = qtx.DeleteAllFeedsInNewsletter(ctx, dbgen.DeleteAllFeedsInNewsletterParams{
		NewsletterID: id,
		UserID:       userId,
	})
	if err != nil {
		return err
	}

	return qtx.DeleteNewsletter(ctx, dbgen.DeleteNewsletterParams{
		ID:     id,
		UserID: userId,
	})
}

func (s *NewsletterService) StoreNewsletterIssue(
	ctx context.Context,
	newsletterId, userId string,
	itemIdToToken map[string]string,
) (string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	issueId, err := qtx.StoreNewsletterIssue(ctx, dbgen.StoreNewsletterIssueParams{
		NewsletterID: newsletterId,
		UserID:       userId,
		SentAt:       time.Now(),
	})
	if err != nil {
		return "", err
	}

	items := make([]dbgen.StoreNewsletterIssueItemsParams, 0)
	for itemId, token := range itemIdToToken {
		items = append(items, dbgen.StoreNewsletterIssueItemsParams{
			ItemID:  itemId,
			UserID:  userId,
			IssueID: issueId,
			Token:   token,
		})
	}

	_, err = qtx.StoreNewsletterIssueItems(ctx, items)
	if err != nil {
		return "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", err
	}

	return issueId, nil
}
