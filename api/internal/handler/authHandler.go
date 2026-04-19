package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/db"
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
	Verified bool    `json:"verified"`
	Tokens   *Tokens `json:"tokens"`
}

type authOutput struct {
	Body authOutputBody
}

type refreshInput struct {
	RefreshToken string `header:"Authorization"`
}

type AuthHandler struct {
	queries      *db.Queries
	emailService email.EmailService
	config       *config.Config
}

func NewAuthHandler(queries *db.Queries, emailService email.EmailService, config *config.Config) *AuthHandler {
	return &AuthHandler{queries, emailService, config}
}

func (h *AuthHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "sign-up",
		Method:        "POST",
		Path:          "/auth/sign-up",
		Summary:       "Sign up",
		DefaultStatus: http.StatusNoContent,
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
		OperationID:   "revoke-token",
		Method:        "POST",
		Path:          "/auth/revoke",
		Summary:       "Revoke a refresh token",
		DefaultStatus: 204,
	}, h.handleRevokeToken)

	huma.Register(api, huma.Operation{
		OperationID:   "resend-eamil-verification",
		Method:        http.MethodPost,
		Path:          "/auth/verify/resend",
		Summary:       "Resend email verification email",
		DefaultStatus: http.StatusNoContent,
	}, h.handleResendVerificationEmail)
}

func (h *AuthHandler) handleSignUp(ctx context.Context, i *authInput) (*authOutput, error) {
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
		if db.IsUniqueViolation(err) {
			return nil, huma.Error409Conflict("Email already in use")
		}

		return nil, internalServerError(ctx, err)
	}
	wideLog.AddLogField(ctx, "didCreateUser", true)

	err = h.sendVerificationEmail(ctx, id, i.Body.Email)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: false,
			Tokens:   nil,
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
		err = h.sendVerificationEmail(ctx, user.ID, user.Email)
		if err != nil {
			return nil, internalServerError(ctx, err)
		}

		return &authOutput{
			Body: authOutputBody{
				Verified: false,
				Tokens:   nil,
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
			Tokens:   tokenResult,
		},
	}, nil
}

func (h *AuthHandler) handleVerifyEmail(ctx context.Context, i *struct {
	Body struct {
		Code  string `json:"code"`
		Email string `json:"email"`
	}
}) (*authOutput, error) {
	user, err := h.queries.GetUserByEmail(ctx, i.Body.Email)
	if err != nil {
		return nil, huma.Error401Unauthorized("Token or email is invalid")
	}

	hashedToken := hashVerificationToken(i.Body.Code)
	_, err = h.queries.FindValidToken(ctx, db.FindValidTokenParams{
		UserID:               user.ID,
		Code:                 hashedToken,
		Purpose:              db.TokenPurposeEmailVerify,
		ExpiresAtGreaterThan: time.Now(),
	})
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return nil, badRequestError("Token or email is invalid.")
		}

		return nil, internalServerError(ctx, err)
	}

	if err = h.queries.DeleteExistingTokensWithPurpose(ctx, db.DeleteExistingTokensWithPurposeParams{
		UserID:  user.ID,
		Purpose: db.TokenPurposeEmailVerify,
	}); err != nil {
		return nil, internalServerError(ctx, err)
	}

	if err = h.queries.MarkUserEmailAsVerified(ctx, user.ID); err != nil {
		return nil, internalServerError(ctx, err)
	}

	tokenResult, err := h.buildTokenResult(ctx, user.ID)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: true,
			Tokens:   tokenResult,
		},
	}, nil
}

func (h *AuthHandler) handleTokenRefresh(ctx context.Context, i *refreshInput) (*authOutput, error) {
	token, err := auth.GetTokenFromAuthorizationHeader(i.RefreshToken)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid authorization header provided")
	}

	tokenData, err := h.queries.GetRefreshToken(ctx, token)
	if err != nil {
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
			Tokens:   tokenResult,
		},
	}, nil
}

func (h *AuthHandler) handleRevokeToken(ctx context.Context, i *refreshInput) (*struct{}, error) {
	token, err := auth.GetTokenFromAuthorizationHeader(i.RefreshToken)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid authorization header provided")
	}

	err = h.queries.RevokeToken(ctx, token)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

func (h *AuthHandler) handleResendVerificationEmail(ctx context.Context, i *struct {
	Body struct {
		Email string `json:"email"`
	}
}) (*struct{}, error) {
	user, err := h.queries.GetUserByEmail(ctx, i.Body.Email)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {

		}
		return nil, huma.Error500InternalServerError(internalServerErrorText)
	}

	if err = h.sendVerificationEmail(ctx, user.ID, user.Email); err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

func (h *AuthHandler) sendVerificationEmail(
	ctx context.Context,
	userID string,
	userEmail string,
) error {
	err := h.queries.DeleteExistingTokensWithPurpose(ctx, db.DeleteExistingTokensWithPurposeParams{
		UserID:  userID,
		Purpose: db.TokenPurposeEmailVerify,
	})
	if err != nil {
		return err
	}

	verificationToken, err := auth.MakeVerificationToken()
	if err != nil {
		return err
	}
	wideLog.AddLogField(ctx, "didMakeVerificationToken", true)

	hashedToken := hashVerificationToken(verificationToken)

	err = h.queries.SaveVerificationToken(ctx, db.SaveVerificationTokenParams{
		UserID:    userID,
		Code:      hashedToken,
		Purpose:   db.TokenPurposeEmailVerify,
		ExpiresAt: time.Now().Add(time.Minute * 10),
	})
	if err != nil {
		return err
	}

	emailHtml, err := h.emailService.AssembleEmail("verify-email.html", map[string]any{
		"Code": verificationToken,
	})
	if err != nil {
		return err
	}

	result, err := h.emailService.Send(
		ctx,
		"Verify your email",
		h.config.NewsletterSenderEmail.Address,
		userEmail,
		emailHtml,
	)
	if err != nil {
		return err
	}

	wideLog.AddLogField(ctx, "emailSendResult", result)
	return nil
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
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	})

	if err != nil {
		return nil, err
	}

	return &Tokens{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func hashVerificationToken(token string) string {
	hashedToken := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashedToken[:])
}
