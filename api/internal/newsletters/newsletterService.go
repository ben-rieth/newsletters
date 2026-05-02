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
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Frequency    string     `json:"frequency"`
	NextSendTime time.Time  `json:"nextSendTime"`
	SendDay      int        `json:"sendDay"`
	SendHour     int        `json:"sendHour"`
	SendMinute   int        `json:"sendMinute"`
	SendTimezone string     `json:"sendTimezone"`
	LastSentAt   *time.Time `json:"lastSentAt,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
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
		return nil, errors.New("Database Failure")
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

	feedsResult, err := service.queries.GetFeedsForManyNewsletters(ctx, newsletterIds)
	if err != nil {
		return nil, errors.New("Database Failure")
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
			Feeds:            feedsByNewsletter[row.ID],
		})
	}

	return &dueNewsletters, nil
}

func (s *NewsletterService) UpdateSendTimes(
	ctx context.Context,
	nl *SendableNewsletter,
	sentAt time.Time,
) error {
	nextSendTime, err := ComputeNextSendTime(
		dbgen.Frequency(nl.Frequency), int(nl.SendDay), int(nl.SendHour), int(nl.SendMinute), nl.SendTimezone,
		time.Now(),
	)
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
