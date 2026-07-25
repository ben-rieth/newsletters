package handler

import (
	"context"

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

	huma.Register(api, huma.Operation{
		OperationID: "get-issues",
		Method:      "GET",
		Path:        "/issues",
		Summary:     "Get all issues for a user",
	}, h.handleGetAllIssues)

	huma.Register(api, huma.Operation{
		OperationID: "get-issue",
		Method:      "GET",
		Path:        "/issues/{id}",
		Summary:     "Get a single issue with all items",
	}, h.handleGetIssue)
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
		issues = append(issues, newsletters.Issue{
			IssueID:        issueRow.ID,
			NewsletterID:   issueRow.NewsletterID,
			NewsletterName: issueRow.Name,
			SentAt:         issueRow.SentAt,
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
		ID string `path:"id"`
	},
) (*getIssueOutput, error) {
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return nil, unauthorizedError()
	}

	issue, err := h.issuesService.GetIssue(ctx, i.ID, claims.Subject)
	if err != nil {
		return nil, internalServerError(ctx, err)
	}

	return &getIssueOutput{
		Body: issue,
	}, nil
}
