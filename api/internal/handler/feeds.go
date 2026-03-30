package handler

import (
	"context"

	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/types"
	"github.com/danielgtaylor/huma/v2"
)

type FeedHandler struct {
	queries *db.Queries
}

func NewFeedHandler(queries *db.Queries) *FeedHandler {
	return &FeedHandler{queries: queries}
}

func (h *FeedHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "add-feed",
		Method: "POST",
		Path: "/newsletters/{newsletterId}/feeds",
		Summary: "Add a feed to a newsletter",
	}, func (ctx context.Context, input *types.AddFeedInput) (*struct{}, error) {
		exists, err := h.queries.DoesNewsletterExist(ctx, input.NewsletterID)

		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to add feed")
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist. Cannot add feed.")
		}
		
		err = h.queries.AddFeed(ctx, db.AddFeedParams{
			NewsletterID: input.NewsletterID,
			Name: input.Body.Name,
			Url: input.Body.Url,
		})

		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to add feed")
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-feeds",
		Method: "GET",
		Path: "/newsletters/{newsletterId}/feeds",
		Summary: "Get all feeds included in a newsletter",
	}, func(ctx context.Context, input *types.ListFeedsInput) (*types.ListFeedsOutput, error) {
		serverError := huma.Error500InternalServerError("Failed to get feeds")
		
		var exists bool
		var err error	

		exists, err = h.queries.DoesNewsletterExist(ctx, input.NewsletterID)

		if err != nil {
			return nil, serverError
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist. Cannot get feeds.")
		}

		var feeds []db.Feed
		feeds, err = h.queries.GetFeedsForNewsletter(ctx, input.NewsletterID)

		if err != nil {
			return nil, serverError
		}

		out := &types.ListFeedsOutput{}
		for _, feed := range feeds {
			out.Body = append(out.Body,dbFeedToFeedType(feed))
		}

		return out, nil
	})
}

func dbFeedToFeedType (feed db.Feed) types.Feed {
	return types.Feed{
		ID: feed.ID,
		NewsletterId: feed.NewsletterID,
		Name: feed.Name,
		Url: feed.Url,
		LastRetrievedAt: &feed.LastRetrievedAt.Time,
		CreatedAt: feed.CreatedAt.Time,
		UpdatedAt: feed.UpdatedAt.Time,
	}
}