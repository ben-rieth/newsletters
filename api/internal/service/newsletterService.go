package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/types"
)

type NewsletterService struct {
	queries *db.Queries
}

func NewNewsletterService (queries *db.Queries) *NewsletterService {
	return &NewsletterService{queries: queries}
}

func (service *NewsletterService) GetDueNewsletters(ctx context.Context) (*[]types.SendableNewsletter, error) {
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

	feedsByNewsletter := make(map[string][]types.BaseFeed)
	for _, row := range feedsResult {
		var lastRetrievedAt time.Time
		if row.LastRetrievedAt.Valid {
			lastRetrievedAt = row.LastRetrievedAt.Time
		} else {
			lastRetrievedAt = lastSendTimeByNewsletter[row.NewsletterID]
		}
		
		feedsByNewsletter[row.NewsletterID] = append(
			feedsByNewsletter[row.NewsletterID], 
			types.BaseFeed{
				Id: row.ID,
				Name: row.Name,
				URL: row.Url,
				LastRetrievedAt: lastRetrievedAt,
			},
		)
	}

	dueNewsletters := make([]types.SendableNewsletter, 0)
	for _, row := range newsletterResult {
		lastSentAt, ok := lastSendTimeByNewsletter[row.ID]

		if !ok {
			continue
		}
		
		dueNewsletters = append(dueNewsletters, types.SendableNewsletter{
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
	nl *types.SendableNewsletter,
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
		LastSentAt: types.ToTimestamp(&sentAt),
	})

	return nil
}