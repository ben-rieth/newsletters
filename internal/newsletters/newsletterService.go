package newsletters

import (
	"context"
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

type NewsletterService struct {
	queries *dbgen.Queries
	db      *pgxpool.Pool
}

func NewNewsletterService(queries *dbgen.Queries, db *pgxpool.Pool) *NewsletterService {
	return &NewsletterService{queries, db}
}

func (service *NewsletterService) GetDueNewsletters(ctx context.Context) (*[]SendableNewsletter, error) {
	newsletterResult, err := service.queries.GetDueNewsletters(ctx)
	if err != nil {
		return nil, err
	}

	newsletterIds := make([]string, 0)
	lastSendTimeByNewsletter := make(map[string]time.Time)
	for _, row := range newsletterResult {
		var lastSentAt time.Time
		if row.LastSentAt.Valid {
			lastSentAt = row.LastSentAt.Time
		} else {
			computedLastSentAt, lastErr := ComputeLastSendTime(
				row.Frequency,
				int(row.SendDay), int(row.SendHour), int(row.SendMinute),
				row.SendTimezone, time.Now(),
			)

			if lastErr != nil {
				wideLog.AddErrorField(ctx, lastErr)
				continue
			}

			lastSentAt = computedLastSentAt
		}

		newsletterIds = append(newsletterIds, row.ID)

		lastSendTimeByNewsletter[row.ID] = lastSentAt
	}

	feedsResult, err := service.queries.GetSendableFeedsForManyNewsletters(ctx, newsletterIds)
	if err != nil {
		return nil, err
	}

	feedsByNewsletter := make(map[string][]feeds.BaseFeed)
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

		feedsByNewsletter[row.NewsletterID] = append(
			feedsByNewsletter[row.NewsletterID],
			baseFeed,
		)
	}

	dueNewsletters := make([]SendableNewsletter, 0)
	for _, row := range newsletterResult {
		lastSentAt, ok := lastSendTimeByNewsletter[row.ID]

		if !ok {
			continue
		}

		dueNewsletters = append(dueNewsletters, SendableNewsletter{
			ID:               row.ID,
			Name:             row.Name,
			Frequency:        string(row.Frequency),
			SendDay:          int(row.SendDay),
			SendHour:         int(row.SendHour),
			SendMinute:       int(row.SendMinute),
			SendTimezone:     row.SendTimezone,
			Email:            row.Email,
			UserID:           row.UserID,
			LastSendTime:     lastSentAt,
			UnsubscribeToken: row.UnsubscribeToken,
			SendWhenEmpty:    row.SendWhenEmpty,
			IsOneOffSend:     row.IsOneOffSend,
			Feeds:            feedsByNewsletter[row.ID],
		})
	}

	return &dueNewsletters, nil
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
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	err = qtx.DeleteAllFeedsInNewsletter(ctx, id)
	if err != nil {
		return err
	}

	err = qtx.DeleteNewsletter(ctx, dbgen.DeleteNewsletterParams{
		ID:     id,
		UserID: userId,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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

func (s *NewsletterService) DeleteIssue(ctx context.Context, issueID, userID string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteItemsForIssue(ctx, dbgen.DeleteItemsForIssueParams{
		IssueID: issueID,
		UserID:  userID,
	}); err != nil {
		return err
	}

	if err := qtx.DeleteIssue(ctx, dbgen.DeleteIssueParams{
		ID:     issueID,
		UserID: userID,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
