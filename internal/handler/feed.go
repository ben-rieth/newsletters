package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	dbutil "github.com/ben-rieth/newsletter-api/internal/db"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
)

type FeedHandler struct {
	queries     *db.Queries
	feedService *feeds.FeedService
}

func NewFeedHandler(queries *db.Queries, feedService *feeds.FeedService) *FeedHandler {
	return &FeedHandler{queries, feedService}
}

type newsletterIdPath struct {
	NewsletterID string `path:"newsletterId"`
}

type feedIdPath struct {
	NewsletterID string `path:"newsletterId"`
	FeedID       string `path:"feedId"`
}

func (h *FeedHandler) RegisterRoutes(api huma.API) {
	type submittableFeedFields struct {
		Alias string `json:"alias"`
		Url   string `json:"url" maxLength:"2048"`
	}

	type addFeedInput struct {
		NewsletterID string `path:"newsletterId"`
		Body         submittableFeedFields
	}

	doesFeedExistMiddleware := newDoesFeedExistMiddleware(api, h.queries)
	doesNewsletterExistMiddleware := newDoesNewsletterExistMiddleware(api, h.queries)

	huma.Register(api, huma.Operation{
		OperationID:   "add-feed",
		Method:        "POST",
		Path:          "/newsletter/{newsletterId}/feed",
		Summary:       "Add a feed to a newsletter",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesNewsletterExistMiddleware},
	}, func(ctx context.Context, input *addFeedInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		feed, err := h.feedService.GetFeedMetaData(ctx, input.Body.Url, true)
		if err != nil {
			return nil, feedFetchError(ctx, err)
		}

		err = h.queries.AddNewsletterFeed(ctx, db.AddNewsletterFeedParams{
			NewsletterID: input.NewsletterID,
			UserID:       claims.Subject,
			Alias:        input.Body.Alias,
			FeedID:       feed.Id,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})

	type uiFeed struct {
		Id          string           `json:"id"`
		Title       string           `json:"title"`
		Description string           `json:"description"`
		Url         string           `json:"url"`
		HtmlURL     string           `json:"htmlUrl"`
		Alias       string           `json:"alias"`
		Health      feeds.FeedHealth `json:"health"`
	}

	type listFeedsOutput struct {
		Body []uiFeed
	}

	huma.Register(api, huma.Operation{
		OperationID: "list-feeds",
		Method:      "GET",
		Path:        "/newsletter/{newsletterId}/feed",
		Summary:     "Get all feeds included in a newsletter",
		Middlewares: huma.Middlewares{doesNewsletterExistMiddleware},
	}, func(ctx context.Context, input *newsletterIdPath) (*listFeedsOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		nlFeeds, err := h.queries.GetFeedsForNewsletter(ctx, input.NewsletterID)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		lastFailures, err := h.queries.GetLastFeedFailuresForNewsletter(
			ctx,
			db.GetLastFeedFailuresForNewsletterParams{
				NewsletterID: input.NewsletterID,
				UserID:       claims.Subject,
				OccurredAt:   feeds.FailureDisplayCutoff(),
			},
		)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		failureByFeed := make(map[string]feeds.FeedFailure, len(lastFailures))
		for _, failure := range lastFailures {
			failureByFeed[failure.NewsletterFeedID] = feeds.FeedFailure{
				OccurredAt: failure.OccurredAt,
				Message:    failure.Message,
			}
		}

		uiFeeds := make([]uiFeed, 0, len(nlFeeds))
		for _, nlFeed := range nlFeeds {
			var lastFailure *feeds.FeedFailure
			if failure, ok := failureByFeed[nlFeed.ID]; ok {
				lastFailure = &failure
			}

			uiFeeds = append(uiFeeds, uiFeed{
				Id:          nlFeed.ID,
				Title:       nlFeed.Title,
				Description: nlFeed.Description,
				Url:         nlFeed.Url,
				HtmlURL:     nlFeed.HtmlUrl,
				Alias:       nlFeed.Alias,
				Health: feeds.BuildFeedHealth(
					dbutil.FromTimestamp(nlFeed.DisabledUntil),
					nlFeed.LastRetrievedAt,
					lastFailure,
				),
			})
		}

		out := &listFeedsOutput{
			Body: uiFeeds,
		}

		return out, nil
	})

	type updatableFeedFields struct {
		Alias string `json:"alias"`
	}

	type updateFeedInput struct {
		NewsletterID string `path:"newsletterId"`
		FeedID       string `path:"feedId"`
		Body         updatableFeedFields
	}

	huma.Register(api, huma.Operation{
		OperationID:   "update-feed",
		Method:        "PUT",
		Path:          "/newsletter/{newsletterId}/feed/{feedId}",
		Summary:       "Update the alias of a feed",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesFeedExistMiddleware},
	}, func(ctx context.Context, input *updateFeedInput) (*struct{}, error) {
		serverError := huma.Error500InternalServerError("Failed to get feeds")

		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := h.queries.UpdateNewsletterFeed(ctx, db.UpdateNewsletterFeedParams{
			NewsletterID: input.NewsletterID,
			ID:           input.FeedID,
			Alias:        input.Body.Alias,
			UserID:       claims.Subject,
		})

		if err != nil {
			return nil, serverError
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-feed",
		Method:        "DELETE",
		Path:          "/newsletter/{newsletterId}/feed/{feedId}",
		Summary:       "Deletes a feed from a newsletter",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesFeedExistMiddleware},
	}, func(ctx context.Context, input *feedIdPath) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := h.queries.DeleteNewsletterFeed(ctx, db.DeleteNewsletterFeedParams{
			NewsletterID: input.NewsletterID,
			ID:           input.FeedID,
			UserID:       claims.Subject,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})

	type uiDetailedFeed struct {
		Id          string             `json:"id"`
		Alias       string             `json:"alias"`
		Title       string             `json:"title"`
		Url         string             `json:"url"`
		HtmlUrl     string             `json:"htmlUrl"`
		Description string             `json:"description"`
		Filters     []feeds.FeedFilter `json:"filters"`
		Health      feeds.FeedHealth   `json:"health"`
	}

	type getFeedOutput struct {
		Body uiDetailedFeed
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-feed",
		Method:      "GET",
		Path:        "/newsletter/{newsletterId}/feed/{feedId}",
		Summary:     "Get data associated with a feed",
		Middlewares: huma.Middlewares{doesFeedExistMiddleware},
	}, func(ctx context.Context, i *feedIdPath) (*getFeedOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		filters, err := h.queries.GetFiltersForFeed(ctx, db.GetFiltersForFeedParams{
			UserID:           claims.Subject,
			NewsletterFeedID: i.FeedID,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		feedFilters := make([]feeds.FeedFilter, 0, len(filters))
		for _, filter := range filters {
			feedFilters = append(feedFilters, feeds.FeedFilter{
				Id:       filter.ID,
				Pattern:  filter.Pattern,
				Operator: filter.Operator,
				Field:    filter.Field,
			})
		}

		feed, err := h.queries.GetFeedById(ctx, db.GetFeedByIdParams{
			ID:     i.FeedID,
			UserID: claims.Subject,
		})
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		var lastFailure *feeds.FeedFailure
		failure, err := h.queries.GetLastFeedFailure(ctx, db.GetLastFeedFailureParams{
			ID:         i.FeedID,
			UserID:     claims.Subject,
			OccurredAt: feeds.FailureDisplayCutoff(),
		})
		if err == nil {
			lastFailure = &feeds.FeedFailure{
				OccurredAt: failure.OccurredAt,
				Message:    failure.Message,
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, internalServerError(ctx, err)
		}

		return &getFeedOutput{
			Body: uiDetailedFeed{
				Id:          i.FeedID,
				Alias:       feed.Alias,
				Title:       feed.Title,
				Url:         feed.Url,
				HtmlUrl:     feed.HtmlUrl,
				Description: feed.Description,
				Filters:     feedFilters,
				Health: feeds.BuildFeedHealth(
					dbutil.FromTimestamp(feed.DisabledUntil),
					feed.LastRetrievedAt,
					lastFailure,
				),
			},
		}, nil
	})

	type existingRecievedFeed struct {
		Alias          string `json:"alias"`
		NewsletterName string `json:"newsletterName"`
	}

	type getFeedMetaDataOutputBody struct {
		Metadata              *feeds.FeedMetaData     `json:"metadata"`
		ExistingRecievedFeeds []*existingRecievedFeed `json:"existingRecievedFeeds"`
	}

	type getFeedMetaDataOutput struct {
		Body *getFeedMetaDataOutputBody
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-feed-metadata",
		Method:      "POST",
		Path:        "/feed",
		Summary:     "Gets data about an RSS feed URL",
	}, func(
		ctx context.Context,
		i *struct {
			Body struct {
				URL string `json:"url" maxLength:"2048"`
			}
		},
	) (*getFeedMetaDataOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := feeds.IsSafeFeedUrl(i.Body.URL)
		if err != nil {
			return nil, badRequestError("Invalid Url")
		}

		metadata, err := h.feedService.GetFeedMetaData(ctx, i.Body.URL, false)
		if err != nil {
			return nil, feedFetchError(ctx, err)
		}

		if metadata == nil {
			return nil, internalServerError(ctx, errors.New("feed metadata was empty"))
		}

		existingUserFeedsResult, err := h.queries.DoesUserAlreadyRecieveFeed(ctx, db.DoesUserAlreadyRecieveFeedParams{
			UserID: claims.Subject,
			FeedID: metadata.Id,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		existingRecievedFeeds := make([]*existingRecievedFeed, 0, len(existingUserFeedsResult))
		for _, existing := range existingUserFeedsResult {
			existingRecievedFeeds = append(existingRecievedFeeds, &existingRecievedFeed{
				NewsletterName: existing.Name,
				Alias:          existing.Alias,
			})
		}

		return &getFeedMetaDataOutput{
			Body: &getFeedMetaDataOutputBody{
				Metadata:              metadata,
				ExistingRecievedFeeds: existingRecievedFeeds,
			},
		}, nil
	})

	type itemPreview struct {
		Id            string            `json:"id"`
		Title         string            `json:"title"`
		Url           string            `json:"url"`
		MatchedFilter *feeds.FeedFilter `json:"matchedFilter,omitempty"`
	}

	type previewFeedOutput struct {
		Body []itemPreview
	}

	huma.Register(api, huma.Operation{
		OperationID: "preview-feed",
		Method:      "GET",
		Path:        "/newsletter/{newsletterId}/feed/{feedId}/preview",
		Summary:     "Preview how filters impact the items in the feed",
		Middlewares: huma.Middlewares{doesFeedExistMiddleware},
	}, func(ctx context.Context, i *feedIdPath) (*previewFeedOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		preview, err := h.queries.PreviewFeed(ctx, db.PreviewFeedParams{
			UserID:                 claims.Subject,
			NewsletterFeedID:       i.FeedID,
			NewsletterID:           i.NewsletterID,
			PublishDateGreaterThan: thirtyDaysAgo,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		itemPreviews := make([]itemPreview, 0, len(preview))

		for _, item := range preview {
			itemPreview := itemPreview{
				Id:    item.ItemID,
				Title: item.Title,
				Url:   item.Url,
			}

			if item.FilterID.Valid {
				matchedFilter := feeds.FeedFilter{
					Id:       item.FilterID.String(),
					Field:    item.Field.FilterField,
					Operator: item.Operator.FilterOperator,
					Pattern:  item.Pattern.String,
				}

				itemPreview.MatchedFilter = &matchedFilter
			}

			itemPreviews = append(itemPreviews, itemPreview)
		}

		return &previewFeedOutput{
			Body: itemPreviews,
		}, nil
	})
}

func newDoesFeedExistMiddleware(api huma.API, queries *db.Queries) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		claims, ok := auth.ClaimsFromContext(ctx.Context())
		if !ok || claims == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, unauthorizedErrorText)
			return
		}

		newsletterId := ctx.Param("newsletterId")
		feedId := ctx.Param("feedId")

		if feedId == "" || newsletterId == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "Request does not have all required information")
			return
		}

		exists, err := queries.DoesFeedExist(ctx.Context(), db.DoesFeedExistParams{
			NewsletterID: newsletterId,
			ID:           feedId,
			UserID:       claims.Subject,
		})

		if err != nil {
			wideLog.AddErrorField(ctx.Context(), err)
			huma.WriteErr(api, ctx, http.StatusInternalServerError, internalServerErrorText)
			return
		}

		if !exists {
			huma.WriteErr(api, ctx, http.StatusNotFound, "This feed does not exist")
			return
		}

		next(ctx)
	}
}
