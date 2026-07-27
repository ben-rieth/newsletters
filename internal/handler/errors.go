package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/utils"
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
	wideLog.AddErrorField(ctx, err)
	return huma.Error500InternalServerError(internalServerErrorText)
}

func feedFetchError(ctx context.Context, err error) huma.StatusError {
	// Every fetch failure is about the feed the caller asked for, not about us,
	// so report the specific reason rather than a 500 — even for the kinds that
	// wrap SystemError.
	var fetchErr *feeds.FetchError
	if errors.As(err, &fetchErr) {
		wideLog.AddErrorField(ctx, err)
		return badRequestError(feeds.DescribeFetchError(err))
	}

	if errors.Is(err, utils.UserError) {
		return badRequestError(feeds.DescribeFetchError(err))
	}

	return internalServerError(ctx, err)
}
