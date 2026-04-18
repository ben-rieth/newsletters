package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
)

type listNewslettersOutput struct {
	Body []newsletters.Newsletter
}

type submittableNewsletterFields struct {
	Name string `json:"name" minLength:"1"`
	Frequency string `json:"frequency" enum:"daily,weekly,monthly"`
	SendDay *int `json:"sendDay,omitempty" minimum:"0" maximum:"31"`
	SendHour int `json:"sendHour" minimum:"0" maximum:"23"`
	SendMinute int `json:"sendMinute" minimum:"0" maximum:"59"`
	SendTimezone string `json:"sendTimezone"`
}

type createNewsletterInput struct {
	Body submittableNewsletterFields
}

type baseNewsletterInput struct {
	ID string `path:"id"`
}

type getNewsletterOutput struct {
    Body newsletters.Newsletter
}

type NewsletterHandler struct {
	queries *db.Queries
	nlService *newsletters.NewsletterService
}

func NewNewsletterHandler(queries *db.Queries, nlService *newsletters.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{queries, nlService}
}

func (h *NewsletterHandler) RegisterRoutes(api huma.API) {
	doesNewsletterExistMiddleware := newDoesNewsletterExistMiddleware(api, h.queries)
	
	huma.Register(api, huma.Operation{
		OperationID: "list-newsletters",
		Method: "GET",
		Path: "/newsletters",
		Summary: "List all newsletters",
	}, func (ctx context.Context, input *struct{}) (*listNewslettersOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}
		
		nls, err := h.queries.ListNewsletters(ctx, claims.Subject)
		
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		out := &listNewslettersOutput{}

		for _, newsletter := range nls {
			out.Body = append(out.Body, newsletters.DbNewsletterToNewsletterType(newsletter))
		}

		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-newsletter",
		Method: "POST",
		Path: "/newsletters",
		Summary: "Create a new newsletter",
	}, func (ctx context.Context, input *createNewsletterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}
		
		var sendDay int
		if input.Body.SendDay == nil {
			sendDay = 0
		} else {
			sendDay = int(*input.Body.SendDay)
		}

		nextSendTime, nextErr := newsletters.ComputeNextSendTime(
			db.Frequency(input.Body.Frequency),
			sendDay, int(input.Body.SendHour), int(input.Body.SendMinute),
			input.Body.SendTimezone, time.Now(),
		)

		if nextErr != nil {
			return nil, huma.Error400BadRequest("Invalid input")
		}
		
		err := h.queries.CreateNewsletter(ctx, db.CreateNewsletterParams{
			Name: input.Body.Name,
			Frequency: db.Frequency(input.Body.Frequency),
			SendDay: int32(sendDay),
			SendHour:  int32(input.Body.SendHour),
			SendMinute: int32(input.Body.SendMinute),
			SendTimezone: input.Body.SendTimezone,
			NextSendTime: nextSendTime,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to create newsletter")
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-newsletter",
		Method: "GET",
		Path: "/newsletters/{id}",
		Summary: "Get a single newsletter",
	}, func (ctx context.Context, input *baseNewsletterInput) (*getNewsletterOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}
		
		newsletter, err := h.queries.GetNewsletter(ctx, db.GetNewsletterParams{
			ID: input.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, notFoundError("Newsletter")
			}

			log.Printf("error: %v", err)
			return nil, internalServerError(ctx, err)
		}

		return &getNewsletterOutput{
			Body: newsletters.DbNewsletterToNewsletterType(newsletter),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-newsletter",
		Method: "PUT",
		Path: "/newsletters/{id}",
		Summary: "Update a single newsletter",
		DefaultStatus: http.StatusNoContent,
		Middlewares: huma.Middlewares{doesNewsletterExistMiddleware},
	}, func (ctx context.Context, input *struct {
		ID string `path:"id"`
		Body submittableNewsletterFields
	}) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		var sendDay int
		if input.Body.SendDay == nil {
			sendDay = 0
		} else {
			sendDay = int(*input.Body.SendDay)
		}

		nextSendTime, err := newsletters.ComputeNextSendTime(
			db.Frequency(input.Body.Frequency),
			sendDay, int(input.Body.SendHour), int(input.Body.SendMinute),
			input.Body.SendTimezone, time.Now(),
		)

		if err != nil {
			return nil, badRequestError("Cannot update newsletter")
		}

		err = h.queries.UpdateNewsletter(ctx, db.UpdateNewsletterParams{
			Name: input.Body.Name,
			Frequency: db.Frequency(input.Body.Frequency),
			SendDay: int32(sendDay),
			SendHour: int32(input.Body.SendHour),
			SendMinute: int32(input.Body.SendMinute),
			SendTimezone: input.Body.SendTimezone,
			NextSendTime: nextSendTime,
			ID: input.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-newsletter",
		Method: "DELETE",
		Path: "/newsletters/{id}",
		Summary: "Delete a newsletter and all of its feeds",
		DefaultStatus: http.StatusNoContent,
		Middlewares: huma.Middlewares{doesNewsletterExistMiddleware},
	}, func(ctx context.Context, input *baseNewsletterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := h.nlService.DeleteNewsletter(ctx, input.ID)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "force-send-newsletter",
		Method: "PATCH",
		Path: "/newsletter/{id}/send",
		Summary: "Updates the newsletter's next send time to the current time",
		DefaultStatus: http.StatusNoContent,
		Middlewares: huma.Middlewares{doesNewsletterExistMiddleware},
	}, func(ctx context.Context, i *baseNewsletterInput) (*struct {}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := h.queries.UpdateNewsletterSendTime(ctx, db.UpdateNewsletterSendTimeParams{
			NextSendTime: time.Now(),
			ID: i.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-newsletter-status",
		Path: "/newsletter/{id}/status",
		Method: http.MethodPatch,
		Description: "Change the status of a newsletter to active or inactive",
		DefaultStatus: http.StatusNoContent,
		Middlewares: huma.Middlewares{doesNewsletterExistMiddleware},
	}, func(ctx context.Context, i *struct {
		ID string `path:"id"`
		Body struct {
			Status string `json:"status" enum:"active,inactive"`
		}
	}) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError()
		}

		err := h.queries.UpdateNewsletterStatus(ctx, db.UpdateNewsletterStatusParams{
			Status: db.NewsletterStatus(i.Body.Status),
			ID: i.ID,
			UserID: claims.Subject,
		})
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})
}

func newDoesNewsletterExistMiddleware(api huma.API, queries *db.Queries) func (ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		claims, ok := auth.ClaimsFromContext(ctx.Context())
		if !ok || claims == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, unauthorizedErrorText)
			return
		}

		newsletterId := ctx.Param("id");

		if newsletterId == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "Request does not have all required information")
			return
		}

		exists, err := queries.DoesNewsletterExist(ctx.Context(), db.DoesNewsletterExistParams{
			ID: newsletterId,
			UserID: claims.Subject,
		})

		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, internalServerErrorText)
			return
		}

		if !exists {
			huma.WriteErr(api, ctx, http.StatusNotFound, notFoundErrorText("feed"))
			return
		}
		
		next(ctx)
	}
}