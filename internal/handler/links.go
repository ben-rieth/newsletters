package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/config"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LinksHandler struct {
	queries *db.Queries
	config  *config.Config
}

func NewLinksHandler(queries *db.Queries, config *config.Config) *LinksHandler {
	return &LinksHandler{queries, config}
}

func (h *LinksHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "link-track",
		Method:        "GET",
		Path:          "/link/{token}",
		Summary:       "Redirect to a newsletter item after marking an item as read",
		DefaultStatus: http.StatusTemporaryRedirect,
	}, h.handleLinkRedirect)
}

type linkRedirectOutput struct {
	Location string `header:"Location"`
}

func (h *LinksHandler) handleLinkRedirect(
	ctx context.Context,
	i *struct {
		Token string `path:"token"`
	},
) (*linkRedirectOutput, error) {
	errorResponse := &linkRedirectOutput{
		Location: fmt.Sprintf("%s/bad-link", h.config.WebURL),
	}

	if err := uuid.Validate(i.Token); err != nil {
		return errorResponse, nil
	}

	url, err := h.queries.GetIssueItemUrlByToken(ctx, i.Token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errorResponse, nil
		}

		return nil, internalServerError(ctx, err)
	}

	if err = h.queries.MarkIssueItemAsReadWithToken(ctx, i.Token); err != nil {
		wideLog.AddErrorField(ctx, fmt.Errorf("failed to mark item as read: %w", err))
	}

	return &linkRedirectOutput{
		Location: url,
	}, nil
}
