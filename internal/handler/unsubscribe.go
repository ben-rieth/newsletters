package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/config"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/email"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
)

type UnsubscribeHandler struct {
	queries      *db.Queries
	cfg          *config.Config
	emailService email.EmailService
}

func NewUnsubscribeHandler(
	queries *db.Queries,
	cfg *config.Config,
	emailService email.EmailService,
) *UnsubscribeHandler {
	return &UnsubscribeHandler{queries, cfg, emailService}
}

func (h *UnsubscribeHandler) RegisterRoutes(api huma.API) {

	huma.Register(api, huma.Operation{
		OperationID:   "unsubscribe",
		Method:        http.MethodPost,
		Path:          "/unsubscribe",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *struct {
		UnsubscribeToken string `query:"unsubscribeToken"`
	}) (*struct{}, error) {
		wideLog.AddLogField(ctx, "unsubscribeToken", i.UnsubscribeToken)

		if i.UnsubscribeToken == "" {
			return nil, badRequestError("Invalid token")
		}

		nl, err := h.queries.GetNewsletterByUnsubscribeToken(ctx, i.UnsubscribeToken)
		if err != nil {
			if errors.Is(pgx.ErrNoRows, err) {
				return nil, badRequestError("Invalid token")
			}

			return nil, internalServerError(ctx, err)
		}

		wideLog.AddLogField(ctx, "newsletterId", nl.NewsletterID)
		wideLog.AddLogField(ctx, "userId", nl.UserID)

		err = h.queries.DeactivateNewsletterByUnsubscribeToken(ctx, i.UnsubscribeToken)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		htmlString, err := h.emailService.AssembleEmail("unsubscribe-receipt.html", map[string]any{
			"NewsletterName": nl.Name,
			"DashboardURL":   fmt.Sprintf("%s/newsletters", h.cfg.WebURL),
		})

		if err != nil {
			wideLog.AddErrorField(ctx, err)
			return nil, nil
		}

		result, err := h.emailService.Send(
			ctx,
			fmt.Sprintf("Unsubscribed from %s", nl.Name),
			h.cfg.NewsletterSenderEmail.Address,
			nl.Email,
			htmlString,
		)

		if err != nil {
			wideLog.AddErrorField(ctx, err)
			return nil, nil
		}

		wideLog.AddLogField(ctx, "sendId", result.ID)
		wideLog.AddLogField(ctx, "emailSendTime", result.Time)

		return nil, nil
	})
}
