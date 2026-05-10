package handler

import (
	"context"
	"net/http"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/danielgtaylor/huma/v2"
)

type LinksHandler struct {
	queries *db.Queries
}

func NewLinksHandler(queries *db.Queries) *LinksHandler {
	return &LinksHandler{queries}
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
	url, err := h.queries.GetIssueItemUrlByToken(ctx, i.Token)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	if err = h.queries.MarkIssueItemAsRead(ctx, i.Token); err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &linkRedirectOutput{
		Location: url,
	}, nil
}
