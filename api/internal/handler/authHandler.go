package handler

import (
	"context"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/crypto/bcrypt"
)

type authInput struct {
	Body struct {
		Email string `json:"email" doc:"Must be a valid email" pattern:"^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$"`
		Password string `json:"password" minLength:"8" maxLength:"80"`
	}
}

type authOutputBody struct {
	Token string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type authOutput struct {
	Body authOutputBody
}

type refreshInput struct {
	RefreshToken string `header:"Authorization"`
}

type AuthHandler struct {
	queries db.Queries
	config config.Config
}

func NewAuthHandler (queries db.Queries, config config.Config) *AuthHandler {
	return &AuthHandler{queries, config}
}

func (h *AuthHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "sign-up",
		Method: "POST",
		Path: "/auth/sign-up",
		Summary: "Sign up",
	}, func(ctx context.Context, i *authInput) (*authOutput, error) {
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

		return &authOutput{
			Body: authOutputBody {
				Token: token,
			},
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "sign-in",
		Method: "POST",
		Path: "/auth/sign-in",
		Summary: "Sign in",
	}, func(ctx context.Context, i *authInput) (*authOutput, error) {
		serverError := huma.Error500InternalServerError("Sign in failed")
		
		user, err := h.queries.GetUserByEmail(ctx, i.Body.Email)
		if err != nil {
			return nil, huma.Error401Unauthorized("Email or password is incorrect")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(i.Body.Password)); err != nil {
			return nil, huma.Error401Unauthorized("Email or password is incorrect")
		}

		token, err := auth.GenerateToken(user.ID, h.config.JWTSecret)
		if err != nil {
			return nil, serverError
		}

		refreshToken := auth.MakeRefreshToken()
		err = h.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
			Token: refreshToken,
			UserID: user.ID,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		})

		if err != nil {
			return nil, serverError
		}

		return &authOutput{
			Body: authOutputBody {
				Token: token,
				RefreshToken: refreshToken,
			},
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "refresh-token",
		Method: "POST",
		Path: "/auth/refresh",
		Summary: "Refresh auth token",
	}, func(ctx context.Context, i *refreshInput) (*authOutput, error) {
		serverError := huma.Error500InternalServerError("Server failed to refresh token")
		
		token, err := auth.GetTokenFromAuthorizationHeader(i.RefreshToken)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid authorization header provided")
		}

		tokenData, err := h.queries.GetRefreshToken(ctx, token)
		if err != nil {
			return nil, serverError
		}

		if tokenData.ExpiresAt.Before(time.Now()) || tokenData.RevokedAt.Valid {
			return nil, huma.Error401Unauthorized("Session is expired")
		}

		err = h.queries.RevokeToken(ctx, tokenData.Token)
		if err != nil {
			return nil, serverError
		}

		jwt, err := auth.GenerateToken(tokenData.UserID, h.config.JWTSecret)
		if err != nil {
			return nil, serverError
		}

		refreshToken := auth.MakeRefreshToken()
		err = h.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
			Token: refreshToken,
			UserID: tokenData.UserID,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		})
		if err != nil {
			return nil, serverError
		}

		return &authOutput{
			Body: authOutputBody{
				Token: jwt,
				RefreshToken: refreshToken,
			},
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "revoke-token",
		Method: "POST",
		Path: "/auth/revoke",
		Summary: "Revoke a refresh token",
		DefaultStatus: 204,
	}, func(ctx context.Context, i *refreshInput) (*struct {}, error) {
		serverError := huma.Error500InternalServerError("Server failed to revoke token")
		
		token, err := auth.GetTokenFromAuthorizationHeader(i.RefreshToken)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid authorization header provided")
		}

		err = h.queries.RevokeToken(ctx, token)
		if err != nil {
			return nil, serverError
		}

		return nil, nil
	})
}