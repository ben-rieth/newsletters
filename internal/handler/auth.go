package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	dbutil "github.com/ben-rieth/newsletter-api/internal/db"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/email"
	"github.com/ben-rieth/newsletter-api/internal/users"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type authInput struct {
	Body struct {
		Email    string `json:"email" doc:"Must be a valid email" pattern:"^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$"`
		Password string `json:"password" minLength:"8" maxLength:"72"`
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
	userService        *users.UserService
}

func NewAuthHandler(
	queries *db.Queries,
	config *config.Config,
	emailVerifyService *email.EmailVerifyService,
	userService *users.UserService,
) *AuthHandler {
	return &AuthHandler{queries, config, emailVerifyService, userService}
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
	userEmail, err := users.CanonicalizeEmail(i.Body.Email)
	if err != nil {
		return nil, badRequestError("Invalid email")
	}

	exists, err := h.queries.IsWhiteListedEmail(ctx, userEmail)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	if !exists {
		return nil, huma.Error403Forbidden("Email not on whitelist. Please contact site owner for invitation.")
	}

	id, err := h.userService.CreateUser(ctx, userEmail, i.Body.Password)
	if err != nil {
		if errors.Is(err, users.EmailInUseError) || dbutil.IsUniqueViolation(err) {
			return nil, huma.Error409Conflict("Email already in use")
		}

		return nil, internalServerError(ctx, err)
	}
	wideLog.AddLogField(ctx, "didCreateUser", true)

	err = h.emailVerifyService.SendVerificationEmail(ctx, id, userEmail)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: false,
		},
	}, nil
}

// A bcrypt hash of a value nobody knows. Comparing against it on the unknown
// email path keeps sign-in's cost the same whether or not the account exists,
// so response time stops being an account oracle.
var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("password-that-is-never-used"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	return hash
}()

func (h *AuthHandler) handleSignIn(ctx context.Context, i *authInput) (*authOutput, error) {
	userEmail, err := users.CanonicalizeEmail(i.Body.Email)
	if err != nil {
		return nil, huma.Error401Unauthorized("Email or password is incorrect")
	}

	user, err := h.queries.GetUserByEmail(ctx, userEmail)
	if err != nil {
		bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(i.Body.Password))
		return nil, huma.Error401Unauthorized("Email or password is incorrect")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(i.Body.Password)); err != nil {
		return nil, huma.Error401Unauthorized("Email or password is incorrect")
	}

	if !user.EmailVerifiedAt.Valid {
		err = h.emailVerifyService.SendVerificationEmail(ctx, user.ID, user.Email)
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
	userEmail, err := users.CanonicalizeEmail(i.Body.Email)
	if err != nil {
		return nil, badRequestError("Invalid email")
	}

	user, err := h.queries.GetUserByEmail(ctx, userEmail)
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
	tokenData, err := h.queries.GetRefreshToken(ctx, auth.HashRefreshToken(i.RefreshToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error401Unauthorized("Invalid token. Please sign in again.")
		}
		return nil, internalServerError(ctx, err)
	}

	// Rotation means a revoked token should never come back. That it did means
	// two parties hold tokens from this chain and there is no way to tell which
	// one is the legitimate user, so the whole family goes.
	if tokenData.RevokedAt.Valid {
		wideLog.AddLogField(ctx, "refreshTokenReuse", true)
		if err := h.queries.DeleteAllRefreshTokensForUser(ctx, tokenData.UserID); err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, huma.Error401Unauthorized("Session is expired")
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
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

	err := h.queries.RevokeToken(ctx, auth.HashRefreshToken(i.RefreshToken))
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
	userEmail, err := users.CanonicalizeEmail(i.Body.Email)
	if err != nil {
		return nil, nil
	}

	user, err := h.queries.GetUserByEmail(ctx, userEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, huma.Error500InternalServerError(internalServerErrorText)
	}

	if user.EmailVerifiedAt.Valid {
		return nil, nil
	}

	if err = h.emailVerifyService.SendVerificationEmail(ctx, user.ID, user.Email); err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

func (h *AuthHandler) buildTokenResult(ctx context.Context, userID string) (*Tokens, error) {
	token, err := auth.GenerateToken(userID, h.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return nil, err
	}

	err = h.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		Token:     auth.HashRefreshToken(refreshToken),
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
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   int(auth.AccessTokenTTL.Seconds()),
		},
		{
			Name:     "refresh_token",
			Value:    tokens.RefreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   int(auth.RefreshTokenTTL.Seconds()),
		},
		{
			Name:     "signed_in",
			Value:    "1",
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
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
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   -1,
		},
		{
			Name:     "refresh_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   -1,
		},
		{
			Name:     "signed_in",
			Value:    "0",
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
			Path:     "/",
		},
	}
}
