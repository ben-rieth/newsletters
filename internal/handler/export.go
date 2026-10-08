package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/danielgtaylor/huma/v2"
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
		Summary:     "Downloads all newsletters, or the given subset, in json format",
	}, func(ctx context.Context, i *struct {
		Ids []string `query:"ids" maxItems:"100" uniqueItems:"true" format:"uuid" doc:"Only export these newsletters; omit to export all"`
	}) (*huma.StreamResponse, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		nls, err := h.listExportable(ctx, claims.Subject, i.Ids)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		if len(i.Ids) > 0 && len(nls) != len(i.Ids) {
			return nil, notFoundError("Newsletter")
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
			exportableNls = append(exportableNls, newsletters.DbNewsletterToExportable(nl, feedsByNl[nl.ID]))
		}

		export := newsletters.NewslettersExport{
			Version:     newsletters.ExportVersion,
			ExportedAt:  time.Now(),
			Newsletters: exportableNls,
		}

		return &huma.StreamResponse{
			Body: func(ctx huma.Context) {
				filename := exportFilename(i.Ids)
				ctx.SetHeader("Content-Type", "application/octet-stream")
				ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

				writer := ctx.BodyWriter()
				json.NewEncoder(writer).Encode(export)
			},
		}, nil
	})
}

func exportFilename(ids []string) string {
	switch len(ids) {
	case 0:
		return "all-newsletters.json"
	case 1:
		return fmt.Sprintf("newsletter-%s.json", ids[0])
	default:
		return "newsletters-export.json"
	}
}

func (h *ExportHander) listExportable(ctx context.Context, userId string, ids []string) ([]db.Newsletter, error) {
	if len(ids) == 0 {
		return h.queries.ListNewsletters(ctx, userId)
	}

	return h.queries.ListNewslettersByIds(ctx, db.ListNewslettersByIdsParams{
		UserID: userId,
		Ids:    ids,
	})
}

func (h *ExportHander) getFeedsByNl(
	ctx context.Context,
	userId string,
	nlIds []string,
) (map[string][]feeds.ExportableFeed, error) {
	fds, err := h.queries.GetFeedsForManyNewsletters(ctx, db.GetFeedsForManyNewslettersParams{
		NewsletterIds: nlIds,
		UserID:        userId,
	})
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

	filtersByFeed := make(map[string][]feeds.ExportableFilter)
	for _, filter := range filters {
		filtersByFeed[filter.NewsletterFeedID] = append(
			filtersByFeed[filter.NewsletterFeedID],
			feeds.ExportableFilter{
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
				Status:   string(feed.Status),
				Filters:  filtersByFeed[feed.NewsletterFeedID],
			},
		)
	}

	return feedsByNl, nil
}
