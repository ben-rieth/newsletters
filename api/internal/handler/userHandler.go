package handler

import (
	"context"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/danielgtaylor/huma/v2"
)

type UserHandler struct {
	queries *db.Queries
}

func NewUserHandler(queries *db.Queries) *UserHandler {
	return &UserHandler{queries}
}

type visibleUser struct {
	Email string `json:"email"`
}

type getUserOutput struct {
	Body visibleUser
}

func (h *UserHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-user",
		Method: "GET",
		Path: "/user",
		Summary: "Get's the user's account data",
	}, func(ctx context.Context, i *struct{}) (*getUserOutput, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}

		user, err := h.queries.GetUserById(ctx, claims.Subject)
		if err != nil {
			return nil, huma.Error500InternalServerError("Server failed to get user data")
		}

		return &getUserOutput{
			Body: visibleUser{
				Email: user.Email,
			},
		}, nil
	})
}