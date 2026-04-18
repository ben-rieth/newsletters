package handler

import (
	"context"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/users"
	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	queries *db.Queries
	userService *users.UserService
}

func NewUserHandler(queries *db.Queries, userService *users.UserService) *UserHandler {
	return &UserHandler{queries, userService}
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
			return nil, internalServerError(ctx, err)
		}

		return &getUserOutput{
			Body: visibleUser{
				Email: user.Email,
			},
		}, nil
	})

	type updateEmailInput struct {
		Body struct {
			Email string `json:"email"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "update-email",
		Method: "PATCH",
		Path: "/user/email",
		Summary: "Update user's email",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *updateEmailInput) (*struct{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}

		err := h.queries.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
			Email: i.Body.Email,
			ID: claims.Subject,
		})

		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})

	type updatePasswordInput struct {
		Body struct {
			CurrentPassword string `json:"currentPassword"`
			NewPassword string `json:"newPassword"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "update-password",
		Method: "PATCH",
		Path: "/user/password",
		Summary: "Update user's password",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *updatePasswordInput) (*struct {}, error) {
		
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}

		user, err := h.queries.GetUserById(ctx, claims.Subject)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(i.Body.CurrentPassword)); err != nil {
			return nil, huma.Error401Unauthorized("Current password is incorrect")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(i.Body.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		err = h.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
			ID: claims.Subject,
			Password: string(hash),
		})
		if err != nil {
			return nil,  internalServerError(ctx, err)
		}

		return nil, nil
	})

	type deleteUserRequest struct {
		Body struct {
			Password string `json:"password"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "delete-user",
		Method: "DELETE",
		Path: "/user",
		Summary: "Delete this user and all of their data",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *deleteUserRequest) (*struct {}, error) {
		
		claims, ok := auth.ClaimsFromContext(ctx)
		if !ok || claims == nil {
			return nil, huma.Error401Unauthorized("Not authorized")
		}

		user, err := h.queries.GetUserById(ctx, claims.Subject)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(i.Body.Password)); err != nil {
			return nil, huma.Error401Unauthorized("Current password is incorrect")
		}

		err = h.userService.DeleteUser(ctx, user.ID)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, nil
	})
}
