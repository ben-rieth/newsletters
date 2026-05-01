package handler

import (
	"context"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/danielgtaylor/huma/v2"
)

type SchedulerHandler struct {
	scheduler *newsletters.Scheduler
}

func NewSchedulerHandler(scheduler *newsletters.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{
		scheduler,
	}
}

func (h *SchedulerHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "run-scheduler",
		Method:        "POST",
		Path:          "/scheduler/run",
		Summary:       "Runs the scheduler immediately - for debugging",
		DefaultStatus: http.StatusNoContent,
	}, func(_ context.Context, i *struct{}) (*struct{}, error) {
		h.scheduler.ForcePoll()
		return nil, nil
	})
}
