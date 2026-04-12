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
	"github.com/jackc/pgx/v5/pgxpool"
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

type updateNewsletterInput struct {
	baseNewsletterInput
	Body submittableNewsletterFields
}

type NewsletterHandler struct {
	queries *db.Queries
	db *pgxpool.Pool
}

func NewNewsletterHandler(queries *db.Queries, db *pgxpool.Pool) *NewsletterHandler {
	return &NewsletterHandler{queries: queries, db: db}
}

func (h *NewsletterHandler) RegisterRoutes(api huma.API) {
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
			return nil, huma.Error500InternalServerError("Failed to fetch newsletters")
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
		
		var sendDay int32
		
		if input.Body.SendDay == nil {
			sendDay = 0
		} else {
			sendDay = int32(*input.Body.SendDay)
		}

		nextSendTime, nextErr := newsletters.ComputeNextSendTime(
			db.Frequency(input.Body.Frequency),
			int(*input.Body.SendDay), int(input.Body.SendHour), int(input.Body.SendMinute),
			input.Body.SendTimezone, time.Now(),
		)

		if nextErr != nil {
			return nil, huma.Error400BadRequest("Invalid input")
		}
		
		err := h.queries.CreateNewsletter(ctx, db.CreateNewsletterParams{
			Name: input.Body.Name,
			Frequency: db.Frequency(input.Body.Frequency),
			SendDay: sendDay,
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
			return nil, unauthorizedError
		}
		
		newsletter, err := h.queries.GetNewsletter(ctx, db.GetNewsletterParams{
			ID: input.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, huma.Error404NotFound("Newsletter does not exist")
			}

			log.Printf("error: %v", err)
			return nil, internalServerError
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
	}, func (ctx context.Context, input *updateNewsletterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError
		}
		
		exists, err := h.queries.DoesNewsletterExist(ctx, db.DoesNewsletterExistParams{
			ID: input.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist.")
		}

		nextSendTime, err := newsletters.ComputeNextSendTime(
			db.Frequency(input.Body.Frequency),
			int(*input.Body.SendDay), int(input.Body.SendHour), int(input.Body.SendMinute),
			input.Body.SendTimezone, time.Now(),
		)

		if err != nil {
			return nil, huma.Error400BadRequest("Cannot update newsletter")
		}
		
		err = h.queries.UpdateNewsletter(ctx, db.UpdateNewsletterParams{
			Name: input.Body.Name,
			Frequency: db.Frequency(input.Body.Frequency),
			SendDay: int32(*input.Body.SendDay),
			SendHour: int32(input.Body.SendHour),
			SendMinute: int32(input.Body.SendMinute),
			SendTimezone: input.Body.SendTimezone,
			NextSendTime: nextSendTime,
			ID: input.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to update newsletter")
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete-newsletter",
		Method: "DELETE",
		Path: "/newsletters/{id}",
		Summary: "Delete a newsletter and all of its feeds",
	}, func(ctx context.Context, input *baseNewsletterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError
		}

		exists, err := h.queries.DoesNewsletterExist(ctx, db.DoesNewsletterExistParams{
			ID: input.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist.")
		}

		tx, err := h.db.Begin(ctx)
		if err != nil {
			return nil, internalServerError
		}

		defer tx.Rollback(ctx)

		qtx := h.queries.WithTx(tx)

		err = qtx.DeleteAllFeedsInNewsletter(ctx, input.ID)
		if err != nil {
			return nil, internalServerError
		}

		err = qtx.DeleteNewsletter(ctx, input.ID)
		if err != nil {
			return nil, internalServerError
		}

		err = tx.Commit(ctx)
		if err != nil {
			return nil, internalServerError
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "force-send-newsletter",
		Method: "PATCH",
		Path: "/newsletter/{id}/send",
		Summary: "Updates the newsletter's next send time to the current time",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *baseNewsletterInput) (*struct {}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, unauthorizedError
		}

		exists, err := h.queries.DoesNewsletterExist(ctx, db.DoesNewsletterExistParams{
			ID: i.ID,
			UserID: claims.Subject,
		})
		if err != nil {
			return nil, internalServerError
		}

		if !exists {
			return nil, huma.Error404NotFound("Newsletter does not exist")
		}

		err = h.queries.UpdateNewsletterSendTime(ctx, db.UpdateNewsletterSendTimeParams{
			NextSendTime: time.Now(),
			ID: i.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError
		}

		return nil, nil
	})
}
