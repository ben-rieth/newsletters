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
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type authInput struct {
	Body struct {
		Email    string `json:"email" doc:"Must be a valid email" pattern:"^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$"`
		Password string `json:"password" minLength:"12" maxLength:"72"`
	}
}

// Per account rather than per IP, which an attacker who can rotate addresses
// does not have to respect.
const maxSignInAttempts = 10
const signInLockout = time.Minute * 15

// Reported for every failure including a lockout, since a distinct status or
// message for a locked account would confirm the account exists.
const signInFailedText = "Email or password is incorrect"

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
	RefreshToken string `cookie:"__Host-refresh_token"`
}

type AuthHandler struct {
	queries            *db.Queries
	db                 *pgxpool.Pool
	config             *config.Config
	emailVerifyService *email.EmailVerifyService
	userService        *users.UserService
}

func NewAuthHandler(
	queries *db.Queries,
	pool *pgxpool.Pool,
	config *config.Config,
	emailVerifyService *email.EmailVerifyService,
	userService *users.UserService,
) *AuthHandler {
	return &AuthHandler{queries, pool, config, emailVerifyService, userService}
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

		if errors.Is(err, users.PasswordTooShortError) || errors.Is(err, users.PasswordTooLongError) {
			return nil, badRequestError(err.Error())
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
		return nil, huma.Error401Unauthorized(signInFailedText)
	}

	user, err := h.queries.GetUserByEmail(ctx, userEmail)
	if err != nil {
		bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(i.Body.Password))
		return nil, huma.Error401Unauthorized(signInFailedText)
	}

	if user.SigninLockedUntil.Valid && user.SigninLockedUntil.Time.After(time.Now()) {
		bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(i.Body.Password))
		wideLog.AddLogField(ctx, "signInLockedOut", true)
		return nil, huma.Error401Unauthorized(signInFailedText)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(i.Body.Password)); err != nil {
		if err := h.queries.RecordFailedSignIn(ctx, db.RecordFailedSignInParams{
			MaxAttempts: maxSignInAttempts,
			LockedUntil: time.Now().Add(signInLockout),
			ID:          user.ID,
		}); err != nil {
			return nil, internalServerError(ctx, err)
		}

		return nil, huma.Error401Unauthorized(signInFailedText)
	}

	if err := h.queries.ResetSignInAttempts(ctx, user.ID); err != nil {
		return nil, internalServerError(ctx, err)
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

	tokenResult, err := issueTokens(ctx, h.queries, h.config, user.ID)
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

	tokenResult, err := issueTokens(ctx, h.queries, h.config, user.ID)
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

var errTokenNotClaimable = errors.New("refresh token could not be claimed")

func (h *AuthHandler) handleTokenRefresh(ctx context.Context, i *refreshInput) (*authOutput, error) {
	hashedToken := auth.HashRefreshToken(i.RefreshToken)

	tokenResult, err := h.rotateRefreshToken(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, errTokenNotClaimable) {
			return nil, h.rejectUnclaimableToken(ctx, hashedToken)
		}

		return nil, internalServerError(ctx, err)
	}

	return &authOutput{
		Body: authOutputBody{
			Verified: true,
		},
		SetCookie: buildAuthCookies(tokenResult),
	}, nil
}

// rotateRefreshToken claims the caller's token and issues its replacement in one
// transaction, so of two requests racing with the same token exactly one wins.
//
// The transaction must finish before this returns, and the caller depends on
// that: a failed claim still leaves a row lock behind, and the reuse path writes
// to those same rows on a pooled connection. Reaching it with this transaction
// still open deadlocks the request against itself.
func (h *AuthHandler) rotateRefreshToken(ctx context.Context, hashedToken string) (*Tokens, error) {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	userID, err := qtx.ClaimRefreshToken(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errTokenNotClaimable
		}

		return nil, err
	}

	tokenResult, err := issueTokens(ctx, qtx, h.config, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return tokenResult, nil
}

// rejectUnclaimableToken explains a failed claim. Rotation means a revoked token
// should never come back: that it did means two parties hold tokens from this
// chain and there is no telling which is the real user, so the whole family goes
// along with every access token already issued from it.
func (h *AuthHandler) rejectUnclaimableToken(ctx context.Context, hashedToken string) error {
	tokenData, err := h.queries.GetRefreshToken(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return huma.Error401Unauthorized("Invalid token. Please sign in again.")
		}
		return internalServerError(ctx, err)
	}

	if !tokenData.RevokedAt.Valid {
		return huma.Error401Unauthorized("Session is expired")
	}

	wideLog.AddLogField(ctx, "refreshTokenReuse", true)

	if err := h.queries.DeleteAllRefreshTokensForUser(ctx, tokenData.UserID); err != nil {
		return internalServerError(ctx, err)
	}

	if err := h.queries.InvalidateUserSessions(ctx, tokenData.UserID); err != nil {
		return internalServerError(ctx, err)
	}

	return huma.Error401Unauthorized("Session is expired")
}

func (h *AuthHandler) handleRevokeToken(ctx context.Context, i *refreshInput) (*authOutput, error) {
	if len(i.RefreshToken) == 0 {
		return nil, badRequestError("Invalid token provided")
	}

	userID, err := h.queries.RevokeToken(ctx, auth.HashRefreshToken(i.RefreshToken))
	tokenWasActive := err == nil

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, internalServerError(ctx, err)
	}

	if tokenWasActive {
		// Deliberately account-wide: an access token says only who it belongs to,
		// so there is no way to aim this at one session. Other devices pay one
		// extra refresh and recover on their own refresh tokens.
		if err := h.queries.InvalidateUserSessions(ctx, userID); err != nil {
			return nil, internalServerError(ctx, err)
		}
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

// queries is a parameter so the refresh path can pass its transaction: the new
// token row has to land with the claim of the old one. Committing them separately
// would let a failed insert leave a revoked token with no replacement, which the
// next refresh could not tell apart from theft.
func issueTokens(ctx context.Context, queries *db.Queries, cfg *config.Config, userID string) (*Tokens, error) {
	token, err := auth.GenerateToken(userID, cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return nil, err
	}

	err = queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
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

// The __Host- prefix makes the browser refuse these unless they are Secure,
// Path=/ and carry no Domain, which puts a cookie planted from a neighbouring
// host out of reach.
func buildAuthCookies(tokens *Tokens) []http.Cookie {
	return []http.Cookie{
		{
			Name:     auth.AccessTokenCookie,
			Value:    tokens.Token,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   int(auth.AccessTokenTTL.Seconds()),
		},
		{
			Name:     auth.RefreshTokenCookie,
			Value:    tokens.RefreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   int(auth.RefreshTokenTTL.Seconds()),
		},
		{
			Name:     auth.SignedInCookie,
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
			Name:     auth.AccessTokenCookie,
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   -1,
		},
		{
			Name:     auth.RefreshTokenCookie,
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   -1,
		},
		{
			Name:     auth.SignedInCookie,
			Value:    "",
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
			Path:     "/",
		},
	}
}
