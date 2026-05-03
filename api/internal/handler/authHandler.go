package handler

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	dbutil "github.com/ben-rieth/newsletter-api/internal/db"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/email"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type authInput struct {
	Body struct {
		Email    string `json:"email" doc:"Must be a valid email" pattern:"^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$"`
		Password string `json:"password" minLength:"8" maxLength:"80"`
	}
}

type Tokens struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type authOutputBody struct {
	Verified bool `json:"verified"`
}

type authOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      authOutputBody
}

type refreshInput struct {
	RefreshToken string `cookie:"refresh_token"`
}

type AuthHandler struct {
	queries            *db.Queries
	config             *config.Config
	emailVerifyService *email.EmailVerifyService
}

func NewAuthHandler(
	queries *db.Queries,
	config *config.Config,
	emailVerifyService *email.EmailVerifyService,
) *AuthHandler {
	return &AuthHandler{queries, config, emailVerifyService}
}

func (h *AuthHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "sign-up",
		Method:      "POST",
		Path:        "/auth/sign-up",
		Summary:     "Sign up",
	}, h.handleSignUp)

	huma.Register(api, huma.Operation{
		OperationID: "sign-in",
		Method:      "POST",
		Path:        "/auth/sign-in",
		Summary:     "Sign in",
	}, h.handleSignIn)

	huma.Register(api, huma.Operation{
		OperationID: "verify-email",
		Method:      "POST",
		Path:        "/auth/verify",
		Summary:     "Verify user email to finish creating account",
	}, h.handleVerifyEmail)

	huma.Register(api, huma.Operation{
		OperationID: "refresh-token",
		Method:      "POST",
		Path:        "/auth/refresh",
		Summary:     "Refresh auth token",
	}, h.handleTokenRefresh)

	huma.Register(api, huma.Operation{
		OperationID: "revoke-token",
		Method:      "POST",
		Path:        "/auth/revoke",
		Summary:     "Revoke a refresh token",
	}, h.handleRevokeToken)

	huma.Register(api, huma.Operation{
		OperationID:   "resend-email-verification",
		Method:        http.MethodPost,
		Path:          "/auth/verify/resend",
		Summary:       "Resend email verification email",
		DefaultStatus: http.StatusNoContent,
	}, h.handleResendVerificationEmail)
}

func (h *AuthHandler) handleSignUp(ctx context.Context, i *authInput) (*authOutput, error) {
	if _, err := mail.ParseAddress(i.Body.Email); err != nil {
		return nil, badRequestError("Invalid email")
	}

	exists, err := h.queries.IsWhiteListedEmail(ctx, i.Body.Email)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	if !exists {
		return nil, huma.Error403Forbidden("Email not on whitelist. Please contact site owner for invitation.")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(i.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}
	wideLog.AddLogField(ctx, "didHashPassword", true)

	var id string
	id, err = h.queries.CreateUser(ctx, db.CreateUserParams{
		Email:    i.Body.Email,
		Password: string(hash),
	})

	if err != nil {
		if dbutil.IsUniqueViolation(err) {
			return nil, huma.Error409Conflict("Email already in use")
		}

		return nil, internalServerError(ctx, err)
	}
	wideLog.AddLogField(ctx, "didCreateUser", true)

	err = h.emailVerifyService.SendVerificationEmail(ctx, id, i.Body.Email)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: false,
		},
	}, nil
}

func (h *AuthHandler) handleSignIn(ctx context.Context, i *authInput) (*authOutput, error) {
	user, err := h.queries.GetUserByEmail(ctx, i.Body.Email)
	if err != nil {
		return nil, huma.Error401Unauthorized("Email or password is incorrect")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(i.Body.Password)); err != nil {
		return nil, huma.Error401Unauthorized("Email or password is incorrect")
	}

	if !user.EmailVerifiedAt.Valid {
		err = h.emailVerifyService.SendVerificationEmail(ctx, user.ID, i.Body.Email)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return &authOutput{
			Body: authOutputBody{
				Verified: false,
			},
		}, nil
	}

	tokenResult, err := h.buildTokenResult(ctx, user.ID)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: true,
		},
		SetCookie: buildAuthCookies(tokenResult),
	}, nil
}

func (h *AuthHandler) handleVerifyEmail(ctx context.Context, i *struct {
	Body struct {
		Code  string `json:"code"`
		Email string `json:"email"`
	}
}) (*authOutput, error) {
	if _, err := mail.ParseAddress(i.Body.Email); err != nil {
		return nil, badRequestError("Invalid email")
	}

	user, err := h.queries.GetUserByEmail(ctx, i.Body.Email)
	if err != nil {
		return nil, badRequestError("Token or email is invalid.")
	}

	err = h.emailVerifyService.VerifyUserEmail(ctx, user.ID, i.Body.Code)
	if err != nil {
		if errors.Is(err, email.InvalidTokenError) {
			return nil, badRequestError("Token or email is invalid.")
		}

		return nil, internalServerError(ctx, err)
	}

	tokenResult, err := h.buildTokenResult(ctx, user.ID)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: true,
		},
		SetCookie: buildAuthCookies(tokenResult),
	}, nil
}

func (h *AuthHandler) handleTokenRefresh(ctx context.Context, i *refreshInput) (*authOutput, error) {
	tokenData, err := h.queries.GetRefreshToken(ctx, i.RefreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error401Unauthorized("Invalid token. Please sign in again.")
		}
		return nil, internalServerError(ctx, err)
	}

	if tokenData.ExpiresAt.Before(time.Now()) || tokenData.RevokedAt.Valid {
		return nil, huma.Error401Unauthorized("Session is expired")
	}

	err = h.queries.RevokeToken(ctx, tokenData.Token)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	tokenResult, err := h.buildTokenResult(ctx, tokenData.UserID)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: true,
		},
		SetCookie: buildAuthCookies(tokenResult),
	}, nil
}

func (h *AuthHandler) handleRevokeToken(ctx context.Context, i *refreshInput) (*authOutput, error) {
	if len(i.RefreshToken) == 0 {
		return nil, badRequestError("Invalid token provided")
	}

	err := h.queries.RevokeToken(ctx, i.RefreshToken)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: false,
		},
		SetCookie: clearAuthCookies(),
	}, nil
}

func (h *AuthHandler) handleResendVerificationEmail(ctx context.Context, i *struct {
	Body struct {
		Email string `json:"email"`
	}
}) (*struct{}, error) {
	user, err := h.queries.GetUserByEmail(ctx, i.Body.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, huma.Error500InternalServerError(internalServerErrorText)
	}

	if err = h.emailVerifyService.SendVerificationEmail(ctx, user.ID, i.Body.Email); err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

func (h *AuthHandler) buildTokenResult(ctx context.Context, userID string) (*Tokens, error) {
	token, err := auth.GenerateToken(userID, h.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken := auth.MakeRefreshToken()
	err = h.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userID,
		ExpiresAt: time.Now().Add(auth.RefreshTokenTTL),
	})

	if err != nil {
		return nil, err
	}

	return &Tokens{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func buildAuthCookies(tokens *Tokens) []http.Cookie {
	return []http.Cookie{
		{
			Name:     "access_token",
			Value:    tokens.Token,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   int(auth.AccessTokenTTL.Seconds()),
		},
		{
			Name:     "refresh_token",
			Value:    tokens.RefreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   int(auth.RefreshTokenTTL.Seconds()),
		},
		{
			Name:     "signed_in",
			Value:    "1",
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   int(auth.RefreshTokenTTL.Seconds()),
		},
	}
}

func clearAuthCookies() []http.Cookie {
	return []http.Cookie{
		{
			Name:     "access_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   -1,
		},
		{
			Name:     "refresh_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   -1,
		},
		{
			Name:     "signed_in",
			Value:    "0",
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
			Path:     "/",
		},
	}
}
