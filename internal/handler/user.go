package handler

import (
	"context"
	"errors"
	"net/http"
	"net/mail"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/email"
	"github.com/ben-rieth/newsletter-api/internal/users"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	queries            *db.Queries
	userService        *users.UserService
	emailVerifyService *email.EmailVerifyService
}

func NewUserHandler(
	queries *db.Queries,
	userService *users.UserService,
	emailVerifyService *email.EmailVerifyService,
) *UserHandler {
	return &UserHandler{queries, userService, emailVerifyService}
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
		Method:      "GET",
		Path:        "/user",
		Summary:     "Get's the user's account data",
	}, h.handleGetUser)

	huma.Register(api, huma.Operation{
		OperationID:   "update-password",
		Method:        "PATCH",
		Path:          "/user/password",
		Summary:       "Update user's password",
		DefaultStatus: http.StatusNoContent,
	}, h.handleUpdatePassword)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-user",
		Method:        "DELETE",
		Path:          "/user",
		Summary:       "Delete this user and all of their data",
		DefaultStatus: http.StatusNoContent,
	}, h.handleDeleteUser)

	huma.Register(api, huma.Operation{
		OperationID:   "update-email",
		Method:        "PATCH",
		Path:          "/user/email",
		Summary:       "Update user's email",
		DefaultStatus: http.StatusNoContent,
	}, h.handleEmailUpdate)

	huma.Register(api, huma.Operation{
		OperationID: "verify-email-update",
		Method:      "POST",
		Path:        "/user/email/verify",
		Summary:     "Verify new email",
	}, h.handleVerifyEmailUpdate)

	huma.Register(api, huma.Operation{
		OperationID:   "resend-email-verification-for-update",
		Method:        http.MethodPost,
		Path:          "/user/verify/resend",
		Summary:       "Resend email verification email for email update flow",
		DefaultStatus: http.StatusNoContent,
	}, h.handleResendVerificationEmail)
}

func (h *UserHandler) handleGetUser(ctx context.Context, i *struct{}) (*getUserOutput, error) {
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
}

func (h *UserHandler) handleUpdatePassword(
	ctx context.Context,
	i *struct {
		Body struct {
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
		}
	},
) (*struct{}, error) {
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

	err = h.userService.UpdatePassword(ctx, claims.Subject, i.Body.NewPassword)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

func (h *UserHandler) handleDeleteUser(
	ctx context.Context,
	i *struct {
		Body struct {
			Password string `json:"password"`
		}
	},
) (*struct{}, error) {
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
}

func (h *UserHandler) handleEmailUpdate(ctx context.Context, i *struct {
	Body struct {
		Email string `json:"email"`
	}
}) (*struct{}, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, huma.Error401Unauthorized("Not authorized")
	}

	if _, err := mail.ParseAddress(i.Body.Email); err != nil {
		return nil, badRequestError("Invalid email")
	}

	exists, err := h.queries.IsWhiteListedEmail(ctx, i.Body.Email)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	if !exists {
		return nil, huma.Error403Forbidden("Email not on whitelist. Please contact site owner to get email whitelisted.")
	}

	_, err = h.queries.GetUserByEmail(ctx, i.Body.Email)
	if err == nil {
		return nil, huma.Error409Conflict("Email is already in use.")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, internalServerError(ctx, err)
	}

	err = h.queries.AddPendingEmailUpdate(ctx, db.AddPendingEmailUpdateParams{
		PendingEmail: i.Body.Email,
		ID:           claims.Subject,
	})

	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	err = h.emailVerifyService.SendVerificationEmail(ctx, claims.Subject, i.Body.Email)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

func (h *UserHandler) handleVerifyEmailUpdate(ctx context.Context, i *struct {
	Body struct {
		Code string `json:"code"`
	}
}) (*struct{}, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, huma.Error401Unauthorized("Not authorized")
	}

	err := h.emailVerifyService.VerifyUserEmailUpdate(ctx, claims.Subject, i.Body.Code)
	if err != nil {
		if errors.Is(err, email.InvalidTokenError) {
			return nil, badRequestError("Token or email is invalid.")
		}

		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

func (h *UserHandler) handleResendVerificationEmail(ctx context.Context, i *struct{}) (*struct{}, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, huma.Error401Unauthorized("Not authorized")
	}

	user, err := h.queries.GetUserById(ctx, claims.Subject)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return nil, nil
		}
		return nil, huma.Error500InternalServerError(internalServerErrorText)
	}

	if _, err := mail.ParseAddress(user.PendingEmail); err != nil {
		return nil, badRequestError("User not in email update flow")
	}

	if err = h.emailVerifyService.SendVerificationEmail(ctx, user.ID, user.PendingEmail); err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}
