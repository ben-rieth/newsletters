package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/danielgtaylor/huma/v2"
)

type NewsletterDebugHandler struct {
	queries *db.Queries
}

func NewNewsletterDebugHandler(queries *db.Queries) *NewsletterDebugHandler {
	return &NewsletterDebugHandler{queries}
}

func (h *NewsletterDebugHandler) RegisterRoutes(api huma.API) {
	doesNewsletterExistMiddleware := newDoesNewsletterExistMiddleware(api, h.queries)

	huma.Register(api, huma.Operation{
		OperationID:   "force-send-newsletter",
		Method:        http.MethodPost,
		Path:          "/debug/newsletters/{newsletterId}/send",
		Summary:       "Immediately queue a newsletter to send - for debugging",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesNewsletterExistMiddleware},
	}, h.handleForceSendNewsletter)
}

func (h *NewsletterDebugHandler) handleForceSendNewsletter(ctx context.Context, i *baseNewsletterInput) (*struct{}, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	err := h.queries.ForceSendNewsletter(ctx, db.ForceSendNewsletterParams{
		NextSendTime: time.Now(),
		ID:           i.NewsletterID,
		UserID:       claims.Subject,
	})
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}
