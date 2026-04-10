package handler

import (
	"context"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/types"
	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	queries db.Queries
	config config.Config
}

func NewUserHandler (queries db.Queries, config config.Config) *UserHandler {
	return &UserHandler{queries, config}
}

func (h *UserHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "sign-up",
		Method: "POST",
		Path: "/auth/sign-up",
		Summary: "Sign up",
	}, func(ctx context.Context, i *types.AuthInput) (*types.AuthOutput, error) {
		serverError := huma.Error500InternalServerError("Sign up failed")
		
		hash, err := bcrypt.GenerateFromPassword([]byte(i.Body.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, serverError
		}

		var id string
		id, err = h.queries.CreateUser(ctx, db.CreateUserParams{
			Email: i.Body.Email,
			Password: string(hash),
		})

		if err != nil {
			if db.IsUniqueViolation(err) {
				return nil, huma.Error409Conflict("Email already in use")
			}

			return nil, serverError
		}

		token, err := auth.GenerateToken(id, h.config.JWTSecret)
		if err == nil {
			return nil, serverError
		}

		return &types.AuthOutput{
			Body: types.AuthOutputBody {
				Token: token,
			},
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "sign-in",
		Method: "POST",
		Path: "/auth/sign-in",
		Summary: "Sign in",
	}, func(ctx context.Context, i *types.AuthInput) (*types.AuthOutput, error) {
		serverError := huma.Error500InternalServerError("Sign in failed")
		
		user, err := h.queries.GetUserByEmail(ctx, i.Body.Email)
		if err != nil {
			return nil, huma.Error401Unauthorized("Email or password is incorrect")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(i.Body.Password)); err != nil {
			return nil, huma.Error401Unauthorized("Email or password is incorrect")
		}

		token, err := auth.GenerateToken(user.ID, h.config.JWTSecret)
		if err == nil {
			return nil, serverError
		}

		return &types.AuthOutput{
			Body: types.AuthOutputBody {
				Token: token,
			},
		}, nil
	})
}