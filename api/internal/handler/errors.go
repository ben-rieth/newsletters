package handler

import (
	"context"
	"fmt"

	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/danielgtaylor/huma/v2"
)

func badRequestError(msg string) huma.StatusError {
	return huma.Error400BadRequest(msg)
}

func notFoundErrorText(entityName string) string {
	return fmt.Sprintf("%s does not exist.", entityName)
}

func notFoundError(entityName string) huma.StatusError {
	return huma.Error404NotFound(notFoundErrorText(entityName))
}

var unauthorizedErrorText = "Not authorized"
func unauthorizedError() huma.StatusError {
	return huma.Error401Unauthorized(unauthorizedErrorText)
}

var internalServerErrorText = "Something went wrong on our end. Please try again."

func internalServerError(ctx context.Context, err error) huma.StatusError {
	wideLog.AddLogField(ctx, "error", err)
	return huma.Error500InternalServerError(internalServerErrorText)
}