package service

import (
	"context"
	"errors"

	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/types"
)

type NewsletterService struct {
	queries *db.Queries
}

func NewNewsletterService (queries *db.Queries) *NewsletterService {
	return &NewsletterService{queries: queries}
}

func (service *NewsletterService) GetDueNewsletters(ctx context.Context) (*[]types.NewsletterWithFeeds, error) {
	newsletterResult, err := service.queries.GetDueNewsletters(ctx)
	if err != nil {
		return nil, errors.New("Database Failure")
	}

	newsletterIds := make([]string, 0)
	for _, newsletterRow := range newsletterResult {
		newsletterIds = append(newsletterIds, newsletterRow.ID)
	}

	feedsResult, err := service.queries.GetFeedsForManyNewsletters(ctx, newsletterIds)
	if err != nil {
		return nil, errors.New("Database Failure")
	}

	feedsByNewsletter := make(map[string][]types.BaseFeed)
	for _, row := range feedsResult {
		feedsByNewsletter[row.NewsletterID] = append(
			feedsByNewsletter[row.NewsletterID], 
			types.BaseFeed{
				Id: row.ID,
				Name: row.Name,
				URL: row.Url,
			},
		)
	}

	dueNewsletters := make([]types.NewsletterWithFeeds, 0)
	for _, row := range newsletterResult {
		dueNewsletters = append(dueNewsletters, types.NewsletterWithFeeds{
			ID: row.ID,
			Name: row.Name,
			Frequency: string(row.Frequency),
			SendDay: int(row.SendDay),
			SendHour: int(row.SendHour),
			SendMinute: int(row.SendMinute),
			SendTimezone: row.SendTimezone,
			Feeds: feedsByNewsletter[row.ID],
		})
	}

	return &dueNewsletters, nil
}