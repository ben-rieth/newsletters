package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/danielgtaylor/huma/v2"
)

type FeedHandler struct {
	queries *db.Queries
}

func NewFeedHandler(queries *db.Queries) *FeedHandler {
	return &FeedHandler{queries: queries}
}

type newsletterIdPath struct {
	NewsletterID string `path:"newsletterId"`
}

type feedIdPath struct {
	newsletterIdPath
	FeedID string `path:"feedId"`
}

type submittableFeedFields struct {
	Name string `json:"name" minLength:"1"`
	Url string `json:"url" doc:"Must be a valid HTTPS URL" pattern:"^https:\\/\\/([\\w-]+\\.)+[\\w]{2,}(\\/[\\w\\-.~:/?#\\[\\]@!$&'()*+,;=%]*)?$"`
}

type addFeedInput struct {
	newsletterIdPath
	Body submittableFeedFields
}

type listFeedsOutput struct {
	Body []feeds.Feed
}

type updateFeedInput struct {
	feedIdPath
	Body submittableFeedFields
}


func (h *FeedHandler) RegisterRoutes(api huma.API) {
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
	}, func(ctx context.Context, input *newsletterIdPath) (*listFeedsOutput, error) {
		serverError := huma.Error500InternalServerError("Failed to get feeds")
		
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
			log.Printf("error: %v", err)
			return nil, serverError
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist. Cannot get feeds.")
		}

		var feeds []db.Feed
		feeds, err = h.queries.GetFeedsForNewsletter(ctx, input.NewsletterID)

		if err != nil {
			log.Printf("error: %v", err)
			return nil, serverError
		}

		out := &listFeedsOutput{}
		for _, feed := range feeds {
			out.Body = append(out.Body,dbFeedToFeedType(feed))
		}

		return out, nil
	})

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

		err = h.queries.UpdateFeed(ctx, db.UpdateFeedParams{
			NewsletterID: input.NewsletterID,
			ID: input.FeedID,
			Name: input.Body.Name,
			Url: input.Body.Url,
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

		err = h.queries.DeleteFeed(ctx, db.DeleteFeedParams{
			NewsletterID: input.NewsletterID,
			ID: input.FeedID,
		})

		if err != nil {
			return nil, serverError
		}

		return nil, nil
	})
}

func dbFeedToFeedType (feed db.Feed) feeds.Feed {
	return feeds.Feed{
		ID: feed.ID,
		NewsletterId: feed.NewsletterID,
		Name: feed.Name,
		Url: feed.Url,
		LastRetrievedAt: &feed.LastRetrievedAt.Time,
		CreatedAt: feed.CreatedAt,
		UpdatedAt: feed.UpdatedAt,
	}
}