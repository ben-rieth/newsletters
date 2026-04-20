package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
)

type ExportHander struct {
	queries *db.Queries
}

func NewExportHandler(queries *db.Queries) *ExportHander {
	return &ExportHander{queries}
}

func (h *ExportHander) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "export-newsletters",
		Method:      http.MethodGet,
		Path:        "/export",
		Summary:     "Downloads all newsletters in json format",
	}, func(ctx context.Context, i *struct{}) (*huma.StreamResponse, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		nls, err := h.queries.ListNewsletters(ctx, claims.Subject)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		nlIds := make([]string, 0, len(nls))
		for _, nl := range nls {
			nlIds = append(nlIds, nl.ID)
		}

		feedsByNl, err := h.getFeedsByNl(ctx, claims.Subject, nlIds)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		exportableNls := make([]newsletters.ExportableNewsletter, 0, len(nls))
		for _, nl := range nls {
			exportableNls = append(
				exportableNls,
				newsletters.ExportableNewsletter{
					ID:           nl.ID,
					Name:         nl.Name,
					Frequency:    string(nl.Frequency),
					SendDay:      int(nl.SendDay),
					SendHour:     int(nl.SendHour),
					SendMinute:   int(nl.SendMinute),
					SendTimezone: nl.SendTimezone,
					Feeds:        feedsByNl[nl.ID],
				},
			)
		}

		export := newsletters.NewslettersExport{
			Version:     1,
			ExportedAt:  time.Now(),
			Newsletters: exportableNls,
		}

		return &huma.StreamResponse{
			Body: func(ctx huma.Context) {
				filename := "all-newsletters.json"
				ctx.SetHeader("Content-Type", "application/octet-stream")
				ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

				writer := ctx.BodyWriter()
				json.NewEncoder(writer).Encode(export)
			},
		}, nil
	})

	type exportNewsletterInput struct {
		NewsletterID string `path:"newsletterId"`
	}

	huma.Register(api, huma.Operation{
		OperationID: "export-newsletter",
		Method:      http.MethodGet,
		Path:        "/export/{newsletterId}",
		Summary:     "Downloads a single newsletter in JSON format",
	}, func(ctx context.Context, i *exportNewsletterInput) (*huma.StreamResponse, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		nl, err := h.queries.GetNewsletter(ctx, db.GetNewsletterParams{
			UserID: claims.Subject,
			ID:     i.NewsletterID,
		})
		if err != nil {
			if errors.Is(pgx.ErrNoRows, err) {
				return nil, notFoundError("newsletter")
			}
			return nil, internalServerError(ctx, err)
		}

		nlIds := []string{nl.ID}
		feedsByNl, err := h.getFeedsByNl(ctx, claims.Subject, nlIds)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		export := newsletters.NewslettersExport{
			Version:    1,
			ExportedAt: time.Now(),
			Newsletters: []newsletters.ExportableNewsletter{{
				ID:           nl.ID,
				Name:         nl.Name,
				Frequency:    string(nl.Frequency),
				SendDay:      int(nl.SendDay),
				SendHour:     int(nl.SendHour),
				SendMinute:   int(nl.SendMinute),
				SendTimezone: nl.SendTimezone,
				Feeds:        feedsByNl[nl.ID],
			}},
		}

		return &huma.StreamResponse{
			Body: func(ctx huma.Context) {
				filename := fmt.Sprintf("newsletter-%s.json", nl.ID)
				ctx.SetHeader("Content-Type", "application/octet-stream")
				ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

				writer := ctx.BodyWriter()
				json.NewEncoder(writer).Encode(export)
			},
		}, nil
	})
}

func (h *ExportHander) getFeedsByNl(
	ctx context.Context,
	userId string,
	nlIds []string,
) (map[string][]feeds.ExportableFeed, error) {
	fds, err := h.queries.GetFeedsForManyNewsletters(ctx, nlIds)
	if err != nil {
		return nil, err
	}

	feedIds := make([]string, 0, len(fds))
	for _, feed := range fds {
		feedIds = append(feedIds, feed.NewsletterFeedID)
	}

	filters, err := h.queries.GetFiltersForManyFeeds(ctx, db.GetFiltersForManyFeedsParams{
		UserID:  userId,
		FeedIds: feedIds,
	})
	if err != nil {
		return nil, err
	}

	filtersByFeed := make(map[string][]feeds.FeedFilter)
	for _, filter := range filters {
		filtersByFeed[filter.NewsletterFeedID] = append(
			filtersByFeed[filter.NewsletterFeedID],
			feeds.FeedFilter{
				Id:       filter.ID,
				Field:    filter.Field,
				Operator: filter.Operator,
				Pattern:  filter.Pattern,
			},
		)
	}

	feedsByNl := make(map[string][]feeds.ExportableFeed)
	for _, feed := range fds {
		feedsByNl[feed.NewsletterID] = append(
			feedsByNl[feed.NewsletterID],
			feeds.ExportableFeed{
				ID:       feed.NewsletterFeedID,
				GlobalID: feed.GlobalFeedID,
				Name:     feed.Title,
				Alias:    feed.Alias,
				URL:      feed.Url,
				Filters:  filtersByFeed[feed.NewsletterFeedID],
			},
		)
	}

	return feedsByNl, nil
}
