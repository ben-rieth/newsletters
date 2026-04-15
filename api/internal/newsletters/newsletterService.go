package newsletters

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Newsletter struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Frequency string `json:"frequency"`
	NextSendTime time.Time `json:"nextSendTime"`
	SendDay int `json:"sendDay"`
	SendHour int `json:"sendHour"`
	SendMinute int `json:"sendMinute"`
	SendTimezone string `json:"sendTimezone"`
	LastSentAt *time.Time `json:"lastSentAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NewsletterService struct {
	queries *db.Queries
	db *pgxpool.Pool
}

func NewNewsletterService (queries *db.Queries, db *pgxpool.Pool) *NewsletterService {
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
				log.Printf("Failed to compute last send time: %v", err)
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
		feedsByNewsletter[row.NewsletterID] = append(
			feedsByNewsletter[row.NewsletterID], 
			feeds.BaseFeed{
				GlobalFeedId: row.GlobalFeedID,
				NewsletterFeedId: row.NewsletterFeedID,
				Name: row.Title,
				URL: row.Url,
				LastRetrievedAt: row.LastRetrievedAt,
			},
		)
	}

	dueNewsletters := make([]SendableNewsletter, 0)
	for _, row := range newsletterResult {
		lastSentAt, ok := lastSendTimeByNewsletter[row.ID]

		if !ok {
			continue
		}
		
		dueNewsletters = append(dueNewsletters, SendableNewsletter{
			ID: row.ID,
			Name: row.Name,
			Frequency: string(row.Frequency),
			SendDay: int(row.SendDay),
			SendHour: int(row.SendHour),
			SendMinute: int(row.SendMinute),
			SendTimezone: row.SendTimezone,
			Email: row.Email,
			UserID: row.UserID,
			LastSendTime: lastSentAt,
			Feeds: feedsByNewsletter[row.ID],
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
		db.Frequency(nl.Frequency), int(nl.SendDay), int(nl.SendHour), int(nl.SendMinute), nl.SendTimezone,
		time.Now(),
	)
	if err != nil {
		return err
	}
	
	s.queries.UpdateNewsletterSendTimes(ctx, db.UpdateNewsletterSendTimesParams{
		ID: nl.ID,
		UserID: nl.UserID,
		NextSendTime: nextSendTime,
		LastSentAt: db.ToTimestamp(&sentAt),
	})

	return nil
}

func (s *NewsletterService) DeleteNewsletter(
	ctx context.Context,
	id string,
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

		err = qtx.DeleteNewsletter(ctx, id)
		if err != nil {
			return err
		}

		return tx.Commit(ctx)
}