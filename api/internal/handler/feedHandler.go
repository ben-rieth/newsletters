package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/utils"
	"github.com/danielgtaylor/huma/v2"
)

type FeedHandler struct {
	queries *db.Queries
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
	FeedID string `path:"feedId"`
}

func (h *FeedHandler) RegisterRoutes(api huma.API) {
	type submittableFeedFields struct {
		Alias string `json:"alias"`
		Url string `json:"url"`
	}

	type addFeedInput struct {
		NewsletterID string `path:"newsletterId"`
		Body submittableFeedFields
	}
	
	huma.Register(api, huma.Operation{
		OperationID: "add-feed",
		Method: "POST",
		Path: "/newsletters/{newsletterId}/feeds",
		Summary: "Add a feed to a newsletter",
		DefaultStatus: http.StatusNoContent,
	}, func (ctx context.Context, input *addFeedInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}
		
		exists, err := h.queries.DoesNewsletterExist(ctx, db.DoesNewsletterExistParams{
			ID: input.NewsletterID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist. Cannot add feed.")
		}


		feed, err := h.feedService.GetFeedMetaData(ctx, input.Body.Url, true)
		if err != nil {
			if errors.Is(err, utils.UserError) {
				return nil, huma.Error400BadRequest("URL is invalid")
			} 
			return nil, internalServerError
		}
		
		err = h.queries.AddNewsletterFeed(ctx, db.AddNewsletterFeedParams{
			NewsletterID: input.NewsletterID,
			UserID: claims.Subject,
			Alias: input.Body.Alias,
			FeedID: feed.Id,
		})

		if err != nil {
			return nil, internalServerError
		}

		return nil, nil
	})

	type uiFeed struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Url         string `json:"url"`
		Alias       string `json:"alias"`
	}

	type listFeedsOutput struct {
		Body []uiFeed
	}

	huma.Register(api, huma.Operation{
		OperationID: "list-feeds",
		Method: "GET",
		Path: "/newsletters/{newsletterId}/feeds",
		Summary: "Get all feeds included in a newsletter",
	}, func(ctx context.Context, input *newsletterIdPath) (*listFeedsOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}

		var exists bool
		var err error	

		exists, err = h.queries.DoesNewsletterExist(ctx, db.DoesNewsletterExistParams{
			ID: input.NewsletterID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist. Cannot get feeds.")
		}

		nlFeeds, err := h.queries.GetFeedsForNewsletter(ctx, input.NewsletterID)
		if err != nil {
			return nil, internalServerError
		}

		uiFeeds := make([]uiFeed, 0, len(nlFeeds))
		for _, nlFeed := range nlFeeds {
			uiFeeds = append(uiFeeds, uiFeed{
				Title: nlFeed.Title,
				Description: nlFeed.Description,
				Url: nlFeed.Url,
				Alias: nlFeed.Alias,
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
		FeedID string `path:"feedId"`
		Body updatableFeedFields
	}


	huma.Register(api, huma.Operation{
		OperationID: "update-feed",
		Method: "PUT",
		Path: "/newsletters/{newsletterId}/feeds/{feedId}",
		Summary: "Update the name and URL of a feed",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, input *updateFeedInput) (*struct{}, error) {
		serverError := huma.Error500InternalServerError("Failed to get feeds")
		
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}
		
		exists, err := h.queries.DoesFeedExist(ctx, db.DoesFeedExistParams{
			NewsletterID: input.NewsletterID,
			ID: input.FeedID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, serverError
		}

		if !exists {
			return nil, huma.Error404NotFound("This feed does not exist.")
		}

		err = h.queries.UpdateNewsletterFeed(ctx, db.UpdateNewsletterFeedParams{
			NewsletterID: input.NewsletterID,
			ID: input.FeedID,
			Alias: input.Body.Alias,
		})

		if err != nil {
			return nil, serverError
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-feed",
		Method: "DELETE",
		Path: "/newsletters/{newsletterId}/feeds/{feedId}",
		Summary: "Deletes a feed from a newsletter",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, input *feedIdPath) (*struct{}, error) {
		serverError := huma.Error500InternalServerError("Failed to get feeds")
	
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}
		
		exists, err := h.queries.DoesFeedExist(ctx, db.DoesFeedExistParams{
			NewsletterID: input.NewsletterID,
			ID: input.FeedID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, serverError
		}

		if !exists {
			return nil, huma.Error404NotFound("This feed does not exist.")
		}

		err = h.queries.DeleteNewsletterFeed(ctx, db.DeleteNewsletterFeedParams{
			NewsletterID: input.NewsletterID,
			ID: input.FeedID,
		})

		if err != nil {
			return nil, serverError
		}

		return nil, nil
	})

	type getFeedMetaDataInput struct {
		Body struct {
			URL string `json:"url"`
		}
	}

	type getFeedMetaDataOutput struct {
		Body feeds.FeedMetaData
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-feed-metadata",
		Method: "POST",
		Path: "/feed",
		Summary: "Gets data about an RSS feed URL",
	}, func(ctx context.Context, i *getFeedMetaDataInput) (*getFeedMetaDataOutput, error) {
		err := feeds.IsSafeFeedUrl(i.Body.URL)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid Url", err)
		}
		
		metadata, err := h.feedService.GetFeedMetaData(ctx, i.Body.URL, false)
		if err != nil || metadata == nil {
			return nil, internalServerError
		}

		return &getFeedMetaDataOutput{
			Body: *metadata,
		}, nil
	})
}
