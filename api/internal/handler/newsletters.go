package handler

import (
	"context"

	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/types"
	"github.com/danielgtaylor/huma/v2"
)

type NewsletterHandler struct {
	queries *db.Queries
}

func NewNewsletterHandler(queries *db.Queries) *NewsletterHandler {
	return &NewsletterHandler{queries: queries}
}

func (h *NewsletterHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-newsletters",
		Method: "GET",
		Path: "/newsletters",
		Summary: "List all newsletters",
	}, func (ctx context.Context, input *struct{}) (*types.ListNewslettersOutput, error) {
		newsletters, err := h.queries.ListNewsletters(ctx)
		
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to fetch newsletters")
		}

		out := &types.ListNewslettersOutput{}

		for _, newsletter := range newsletters {
			out.Body = append(out.Body, types.Newsletter{
				ID: newsletter.ID.String(),
				Name: newsletter.Name,
				Frequency: string(newsletter.Frequency),
				SendDay: int(newsletter.SendDay),
				SendHour: int(newsletter.SendHour),
				SendMinute: int(newsletter.SendMinute),
				CreatedAt: newsletter.CreatedAt.Time,
				UpdatedAt: newsletter.UpdatedAt.Time,
			})
		}

		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-newsletter",
		Method: "POST",
		Path: "/newsletters",
		Summary: "Create a new newsletter",
	}, func (ctx context.Context, input *types.CreateNewsletterInput) (*struct{}, error) {
		var sendDay int32
		
		if input.Body.SendDay == nil {
			sendDay = 0
		} else {
			sendDay = int32(*input.Body.SendDay)
		}
		
		err := h.queries.CreateNewsletter(ctx, db.CreateNewsletterParams{
			Name: input.Body.Name,
			Frequency: db.Frequency(input.Body.Frequency),
			SendDay: sendDay,
			SendHour:  int32(input.Body.SendHour),
			SendMinute: int32(input.Body.SendMinute),
		})

		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to create newsletter")
		}

		return nil, nil
	})
}