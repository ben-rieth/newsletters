package handler

import (
	"context"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/danielgtaylor/huma/v2"
)

type IssuesHandler struct {
	queries       *db.Queries
	issuesService *newsletters.IssuesService
}

func NewIssuesHandler(queries *db.Queries, issuesService *newsletters.IssuesService) *IssuesHandler {
	return &IssuesHandler{queries, issuesService}
}

func (h *IssuesHandler) RegisterRoutes(api huma.API) {

	doesIssueExistMiddleware := newDoesIssueExistMiddleware(api, h.queries)
	doesIssueItemExistMiddleware := newDoesIssueItemExistMiddleware(api, h.queries)

	huma.Register(api, huma.Operation{
		OperationID: "get-issues",
		Method:      "GET",
		Path:        "/issues",
		Summary:     "Get all issues for a user",
	}, h.handleGetAllIssues)

	huma.Register(api, huma.Operation{
		OperationID: "get-issue",
		Method:      "GET",
		Path:        "/issues/{issueId}",
		Summary:     "Get a single issue with all items",
	}, h.handleGetIssue)

	huma.Register(api, huma.Operation{
		OperationID:   "update-issue-state",
		Method:        "PUT",
		Path:          "/issues/{issueId}/state",
		Summary:       "Sets the read status of every item in an issue",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesIssueExistMiddleware},
	}, h.handleUpdateIssueState)

	huma.Register(api, huma.Operation{
		OperationID:   "update-issue-item-state",
		Method:        "PUT",
		Path:          "/issues/{issueId}/item/{itemId}/state",
		Summary:       "Updates the read status of a single item in an issue",
		DefaultStatus: http.StatusNoContent,
		Middlewares:   huma.Middlewares{doesIssueItemExistMiddleware},
	}, h.handleUpdateIssueItemState)
}

type allIssuesOutput struct {
	Body []newsletters.Issue
}

func (h *IssuesHandler) handleGetAllIssues(
	ctx context.Context,
	i *struct{},
) (*allIssuesOutput, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	issueRows, err := h.queries.GetAllUserIssues(ctx, claims.Subject)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	issues := make([]newsletters.Issue, 0, len(issueRows))
	for _, issueRow := range issueRows {
		previewTitles := issueRow.PreviewTitles
		if previewTitles == nil {
			previewTitles = []string{}
		}

		issues = append(issues, newsletters.Issue{
			IssueID:        issueRow.ID,
			NewsletterID:   issueRow.NewsletterID,
			NewsletterName: issueRow.Name,
			SentAt:         issueRow.SentAt,
			State:          issueRow.State,
			ItemCount:      issueRow.ItemCount,
			UnreadCount:    issueRow.UnreadCount,
			PreviewTitles:  previewTitles,
		})
	}

	return &allIssuesOutput{
		Body: issues,
	}, nil
}

type getIssueOutput struct {
	Body *newsletters.DetailedIssue
}

func (h *IssuesHandler) handleGetIssue(
	ctx context.Context,
	i *struct {
		IssueID string `path:"issueId" format:"uuid"`
	},
) (*getIssueOutput, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	issue, err := h.issuesService.GetIssue(ctx, i.IssueID, claims.Subject)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &getIssueOutput{
		Body: issue,
	}, nil
}

type issueStateBody struct {
	State db.ItemState `json:"state" enum:"read,unread"`
}

type updateIssueStateInput struct {
	IssueID string `path:"issueId" format:"uuid"`
	Body    issueStateBody
}

func (h *IssuesHandler) handleUpdateIssueState(
	ctx context.Context,
	i *updateIssueStateInput,
) (*struct{}, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	if err := h.queries.UpdateAllIssueItemsState(ctx, db.UpdateAllIssueItemsStateParams{
		IssueID: i.IssueID,
		UserID:  claims.Subject,
		State:   i.Body.State,
	}); err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

type updateIssueItemStateInput struct {
	IssueID string `path:"issueId" format:"uuid"`
	ItemID  string `path:"itemId" format:"uuid"`
	Body    issueStateBody
}

func (h *IssuesHandler) handleUpdateIssueItemState(
	ctx context.Context,
	i *updateIssueItemStateInput,
) (*struct{}, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	if err := h.queries.UpdateIssueItemState(ctx, db.UpdateIssueItemStateParams{
		ItemID:  i.ItemID,
		IssueID: i.IssueID,
		UserID:  claims.Subject,
		State:   i.Body.State,
	}); err != nil {
		return nil, internalServerError(ctx, err)
	}

	return nil, nil
}

func newDoesIssueExistMiddleware(api huma.API, queries *db.Queries) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		claims, ok := auth.ClaimsFromContext(ctx.Context())
		if !ok || claims == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, unauthorizedErrorText)
			return
		}

		issueId := ctx.Param("issueId")

		if issueId == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "Request does not have all required information")
			return
		}

		exists, err := queries.DoesIssueExist(ctx.Context(), db.DoesIssueExistParams{
			ID:     issueId,
			UserID: claims.Subject,
		})

		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, internalServerErrorText)
			return
		}

		if !exists {
			huma.WriteErr(api, ctx, http.StatusNotFound, notFoundErrorText("issue"))
			return
		}

		next(ctx)
	}
}

func newDoesIssueItemExistMiddleware(api huma.API, queries *db.Queries) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		claims, ok := auth.ClaimsFromContext(ctx.Context())
		if !ok || claims == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, unauthorizedErrorText)
			return
		}

		issueId := ctx.Param("issueId")
		issueItemId := ctx.Param("itemId")

		if issueId == "" || issueItemId == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "Request does not have all required information")
			return
		}

		exists, err := queries.DoesIssueItemExist(ctx.Context(), db.DoesIssueItemExistParams{
			ItemID:  issueItemId,
			IssueID: issueId,
			UserID:  claims.Subject,
		})

		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, internalServerErrorText)
			return
		}

		if !exists {
			huma.WriteErr(api, ctx, http.StatusNotFound, notFoundErrorText("issue item"))
			return
		}

		next(ctx)
	}
}
