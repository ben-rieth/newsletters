package handler

import (
	"context"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/danielgtaylor/huma/v2"
)

type FeedFilterHandler struct {
	queries *db.Queries
}

func NewFeedFilterHandler(queries *db.Queries) *FeedFilterHandler {
	return &FeedFilterHandler{queries}
}

func (h *FeedFilterHandler) RegisterRoutes(api huma.API) {
	
	type submittableFeedFilterFields struct {
		Field string `json:"field"`
		Operator string `json:"operator"`
		Pattern string `json:"pattern"`
	}

	type addFilterInput struct {
		NewsletterID string `path:"newsletterId"`
		FeedID string `path:"feedId"`
		Body submittableFeedFilterFields
	}

	huma.Register(api, huma.Operation{
		OperationID: "add-feed-filter",
		Method: "POST",
		Path: "/newsletter/{newsletterId}/feed/{feedId}/filter",
		Summary: "Add a filter to a newsletter feed",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *addFilterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError
		}

		exists, err := h.queries.DoesFeedExist(ctx, db.DoesFeedExistParams{
			UserID: claims.Subject,
			NewsletterID: i.NewsletterID,
			ID: i.FeedID,
		})

		if err != nil {
			return nil, internalServerError
		}

		if !exists {
			return nil, huma.Error404NotFound("This feed does not exist.")
		}

		err = h.queries.AddFeedFilter(ctx, db.AddFeedFilterParams{
			NewsletterFeedID: i.FeedID,
			UserID: claims.Subject,
			Field: db.FilterField(i.Body.Field),
			Operator: db.FilterOperator(i.Body.Operator),
			Pattern: i.Body.Pattern,
		})

		if err != nil {
			return nil, internalServerError
		}

		return nil, nil
	})

	type updateFilterInput struct {
		NewsletterID string `path:"newsletterId"`
		FeedID string `path:"feedId"`
		FilterID string `path:"filterId"`
		Body submittableFeedFilterFields
	}

	huma.Register(api, huma.Operation{
		OperationID: "update-feed-filter",
		Method: "POST",
		Path: "/newsletter/{newsletterId}/feed/{feedId}/filter/{filterId}",
		Summary: "Update a filter to a newsletter feed",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *updateFilterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError
		}

		exists, err := h.queries.DoesFilterExist(ctx, db.DoesFilterExistParams{
			UserID: claims.Subject,
			NewsletterID: i.NewsletterID,
			FilterID: i.FilterID,
			NewsletterFeedID: i.FeedID,
		})

		if err != nil {
			return nil, internalServerError
		}

		if !exists {
			return nil, huma.Error404NotFound("This filter does not exist.")
		}

		err = h.queries.UpdateFeedFilter(ctx, db.UpdateFeedFilterParams{
			ID: i.FilterID,
			Field: db.FilterField(i.Body.Field),
			Operator: db.FilterOperator(i.Body.Operator),
			Pattern: i.Body.Pattern,
		})

		if err != nil {
			return nil, internalServerError
		}

		return nil, nil
	})

	type deleteFilterInput struct {
		NewsletterID string `path:"newsletterId"`
		FeedID string `path:"feedId"`
		FilterID string `path:"filterId"`
	}

	huma.Register(api, huma.Operation{
		OperationID: "delete-feed-filter",
		Method: "DELETE",
		Path: "/newsletter/{newsletterId}/feed/{feedId}/filter/{filterId}",
		Summary: "Delete a filter on a newsletter feed",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *deleteFilterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError
		}

		exists, err := h.queries.DoesFilterExist(ctx, db.DoesFilterExistParams{
			UserID: claims.Subject,
			NewsletterID: i.NewsletterID,
			FilterID: i.FilterID,
			NewsletterFeedID: i.FeedID,
		})

		if err != nil {
			return nil, internalServerError
		}

		if !exists {
			return nil, huma.Error404NotFound("This filter does not exist.")
		}

		err = h.queries.DeleteFeedFilter(ctx, db.DeleteFeedFilterParams{
			ID: i.FilterID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError
		}

		return nil, nil
	})
}