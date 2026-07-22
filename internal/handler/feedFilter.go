package handler

import (
	"context"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
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
		Field    string `json:"field"`
		Operator string `json:"operator"`
		Pattern  string `json:"pattern"`
	}

	type addFilterInput struct {
		NewsletterID string `path:"newsletterId"`
		FeedID       string `path:"feedId"`
		Body         submittableFeedFilterFields
	}

	doesFeedExistMiddleware := newDoesFeedExistMiddleware(api, h.queries)
	doesFilterExistMiddleware := newDoesFilterExistMiddleware(api, h.queries)

	huma.Register(api, huma.Operation{
		OperationID:   "add-feed-filter",
		Method:        "POST",
		Path:          "/newsletter/{newsletterId}/feed/{feedId}/filter",
		Summary:       "Add a filter to a newsletter feed",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesFeedExistMiddleware},
	}, func(ctx context.Context, i *addFilterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := h.queries.AddFeedFilter(ctx, db.AddFeedFilterParams{
			NewsletterFeedID: i.FeedID,
			UserID:           claims.Subject,
			Field:            db.FilterField(i.Body.Field),
			Operator:         db.FilterOperator(i.Body.Operator),
			Pattern:          i.Body.Pattern,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})

	type updateFilterInput struct {
		NewsletterID string `path:"newsletterId"`
		FeedID       string `path:"feedId"`
		FilterID     string `path:"filterId"`
		Body         submittableFeedFilterFields
	}

	huma.Register(api, huma.Operation{
		OperationID:   "update-feed-filter",
		Method:        "POST",
		Path:          "/newsletter/{newsletterId}/feed/{feedId}/filter/{filterId}",
		Summary:       "Update a filter to a newsletter feed",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesFilterExistMiddleware},
	}, func(ctx context.Context, i *updateFilterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := h.queries.UpdateFeedFilter(ctx, db.UpdateFeedFilterParams{
			ID:       i.FilterID,
			Field:    db.FilterField(i.Body.Field),
			Operator: db.FilterOperator(i.Body.Operator),
			Pattern:  i.Body.Pattern,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})

	type deleteFilterInput struct {
		NewsletterID string `path:"newsletterId"`
		FeedID       string `path:"feedId"`
		FilterID     string `path:"filterId"`
	}

	huma.Register(api, huma.Operation{
		OperationID:   "delete-feed-filter",
		Method:        "DELETE",
		Path:          "/newsletter/{newsletterId}/feed/{feedId}/filter/{filterId}",
		Summary:       "Delete a filter on a newsletter feed",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesFilterExistMiddleware},
	}, func(ctx context.Context, i *deleteFilterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := h.queries.DeleteFeedFilter(ctx, db.DeleteFeedFilterParams{
			ID:     i.FilterID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})
}

func newDoesFilterExistMiddleware(api huma.API, queries *db.Queries) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		claims, ok := auth.ClaimsFromContext(ctx.Context())
		if !ok || claims == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, unauthorizedErrorText)
			return
		}

		newsletterId := ctx.Param("newsletterId")
		feedId := ctx.Param("feedId")
		filterId := ctx.Param("filterId")

		if feedId == "" || newsletterId == "" || filterId == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "Request does not have all required information")
			return
		}

		exists, err := queries.DoesFilterExist(ctx.Context(), db.DoesFilterExistParams{
			NewsletterID:     newsletterId,
			FilterID:         filterId,
			NewsletterFeedID: feedId,
			UserID:           claims.Subject,
		})

		if err != nil {
			wideLog.AddErrorField(ctx.Context(), err)
			huma.WriteErr(api, ctx, http.StatusInternalServerError, internalServerErrorText)
			return
		}

		if !exists {
			huma.WriteErr(api, ctx, http.StatusNotFound, notFoundErrorText("Filter"))
			return
		}

		next(ctx)
	}
}
