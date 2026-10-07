package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/danielgtaylor/huma/v2"
)

type ImportHandler struct {
	queries       *db.Queries
	importService *newsletters.ImportService
}

func NewImportHandler(queries *db.Queries, importService *newsletters.ImportService) *ImportHandler {
	return &ImportHandler{queries, importService}
}

type importNewslettersInput struct {
	Body newsletters.NewslettersExport
}

type importNewslettersOutput struct {
	Body newsletters.ImportResult
}

type feedImport struct {
	Id    string             `json:"id"`
	Url   string             `json:"url"`
	Alias string             `json:"alias"`
	State db.FeedImportState `json:"state" enum:"pending,failed"`
	Error string             `json:"error"`
}

type listFeedImportsOutput struct {
	Body []feedImport
}

type feedImportPath struct {
	NewsletterID string `path:"newsletterId"`
	ImportID     string `path:"importId" format:"uuid"`
}

func (h *ImportHandler) RegisterRoutes(api huma.API) {
	doesNewsletterExistMiddleware := newDoesNewsletterExistMiddleware(api, h.queries)

	huma.Register(api, huma.Operation{
		OperationID: "import-newsletters",
		Method:      http.MethodPost,
		Path:        "/import",
		Summary:     "Creates newsletters from a JSON export; feeds are added in the background",
	}, h.handleImportNewsletters)

	huma.Register(api, huma.Operation{
		OperationID: "list-feed-imports",
		Method:      http.MethodGet,
		Path:        "/newsletter/{newsletterId}/feed-imports",
		Middlewares: huma.Middlewares{doesNewsletterExistMiddleware},
	}, h.handleListFeedImports)

	huma.Register(api, huma.Operation{
		OperationID:   "retry-feed-import",
		Method:        http.MethodPost,
		Path:          "/newsletter/{newsletterId}/feed-imports/{importId}/retry",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesNewsletterExistMiddleware},
	}, h.handleRetryFeedImport)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-feed-import",
		Method:        http.MethodDelete,
		Path:          "/newsletter/{newsletterId}/feed-imports/{importId}",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesNewsletterExistMiddleware},
	}, h.handleDeleteFeedImport)
}

func (h *ImportHandler) handleImportNewsletters(ctx context.Context, input *importNewslettersInput) (*importNewslettersOutput, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	result, err := h.importService.Import(ctx, claims.Subject, input.Body)
	if errors.Is(err, newsletters.ErrInvalidImport) {
		return nil, badRequestError(err.Error())
	}
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &importNewslettersOutput{Body: *result}, nil
}

func (h *ImportHandler) handleListFeedImports(ctx context.Context, input *newsletterIdPath) (*listFeedImportsOutput, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	rows, err := h.queries.GetFeedImportsForNewsletter(ctx, db.GetFeedImportsForNewsletterParams{
		NewsletterID: input.NewsletterID,
		UserID:       claims.Subject,
	})
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	imports := make([]feedImport, 0, len(rows))
	for _, row := range rows {
		imports = append(imports, feedImport{
			Id:    row.ID,
			Url:   row.Url,
			Alias: row.Alias,
			State: row.State,
			Error: row.Error,
		})
	}

	return &listFeedImportsOutput{Body: imports}, nil
}

func (h *ImportHandler) handleRetryFeedImport(ctx context.Context, input *feedImportPath) (*struct{}, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	retried, err := h.importService.RetryFeedImport(ctx, input.ImportID, input.NewsletterID, claims.Subject)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}
	if !retried {
		return nil, notFoundError("Failed feed import")
	}

	return nil, nil
}

func (h *ImportHandler) handleDeleteFeedImport(ctx context.Context, input *feedImportPath) (*struct{}, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	deleted, err := h.queries.DeleteFeedImport(ctx, db.DeleteFeedImportParams{
		ID:           input.ImportID,
		NewsletterID: input.NewsletterID,
		UserID:       claims.Subject,
	})
	if err != nil {
		return nil, internalServerError(ctx, err)
	}
	if deleted == 0 {
		return nil, notFoundError("Feed import")
	}

	return nil, nil
}
