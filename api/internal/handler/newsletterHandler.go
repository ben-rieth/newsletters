package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/service"
	"github.com/ben-rieth/newsletter-api/internal/types"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	}, func (ctx context.Context, input *struct{}) (*types.ListNewslettersOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}
		
		newsletters, err := h.queries.ListNewsletters(ctx, claims.Subject)
		
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to fetch newsletters")
		}

		out := &types.ListNewslettersOutput{}

		for _, newsletter := range newsletters {
			out.Body = append(out.Body, dbNewsletterToNewsletterType(newsletter))
		}

		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-newsletter",
		Method: "POST",
		Path: "/newsletters",
		Summary: "Create a new newsletter",
	}, func (ctx context.Context, input *types.CreateNewsletterInput) (*struct{}, error) {
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

		nextSendTime, nextErr := service.ComputeNextSendTime(
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
	}, func (ctx context.Context, input *types.BaseNewsletterInput) (*types.GetNewsletterOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
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
			return nil, huma.Error500InternalServerError("Could not get newsletter")
		}

		return &types.GetNewsletterOutput{
			Body: dbNewsletterToNewsletterType(newsletter),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-newsletter",
		Method: "PUT",
		Path: "/newsletters/{id}",
		Summary: "Update a single newsletter",
	}, func (ctx context.Context, input *types.UpdateNewsletterInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}
		
		exists, err := h.queries.DoesNewsletterExist(ctx, db.DoesNewsletterExistParams{
			ID: input.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to update newsletter")
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist.")
		}

		nextSendTime, err := service.ComputeNextSendTime(
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
	}, func(ctx context.Context, input *types.BaseNewsletterInput) (*struct{}, error) {
		serverError := huma.Error500InternalServerError("Failed to delete newsletter")

		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}

		exists, err := h.queries.DoesNewsletterExist(ctx, db.DoesNewsletterExistParams{
			ID: input.ID,
			UserID: claims.Subject,
		})

		if err != nil {
			return nil, serverError
		}

		if !exists {
			return nil, huma.Error400BadRequest("Newsletter does not exist.")
		}

		tx, err := h.db.Begin(ctx)
		if err != nil {
			return nil, serverError
		}

		defer tx.Rollback(ctx)

		qtx := h.queries.WithTx(tx)

		err = qtx.DeleteAllFeedsInNewsletter(ctx, input.ID)
		if err != nil {
			return nil, serverError
		}

		err = qtx.DeleteNewsletter(ctx, input.ID)
		if err != nil {
			return nil, serverError
		}

		err = tx.Commit(ctx)
		if err != nil {
			return nil, serverError
		}

		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "force-send-newsletter",
		Method: "PATCH",
		Path: "/newsletter/{id}/send",
		Summary: "Updates the newsletter's next send time to the current time",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *types.BaseNewsletterInput) (*struct {}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}

		exists, err := h.queries.DoesNewsletterExist(ctx, db.DoesNewsletterExistParams{
			ID: i.ID,
			UserID: claims.Subject,
		})
		if err != nil {
			return nil, types.InternalServerError
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
			return nil, types.InternalServerError
		}

		return nil, nil
	})
}

func dbNewsletterToNewsletterType (newsletter db.Newsletter) types.Newsletter {
	var lastSentAt *time.Time
	if newsletter.LastSentAt.Valid {
		lastSentAt = &newsletter.LastSentAt.Time
	} else {
		lastSentAt = nil
	}
	
	return types.Newsletter{
		ID: newsletter.ID,
		Name: newsletter.Name,
		Frequency: string(newsletter.Frequency),
		SendDay: int(newsletter.SendDay),
		SendHour: int(newsletter.SendHour),
		SendMinute: int(newsletter.SendMinute),
		SendTimezone: newsletter.SendTimezone,
		LastSentAt: lastSentAt,
		NextSendTime: newsletter.NextSendTime,
		CreatedAt: newsletter.CreatedAt,
		UpdatedAt: newsletter.UpdatedAt,
	}
}