package handler

import (
	"context"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/scheduler"
	"github.com/danielgtaylor/huma/v2"
)

type SchedulerHandler struct {
	scheduler *scheduler.Scheduler
}

func NewSchedulerHandler (scheduler *scheduler.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{
		scheduler,
	}
}

func (h *SchedulerHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "run-scheduler",
		Method: "POST",
		Path: "/scheduler/run",
		Summary: "Runs the scheduler immediately - for debugging",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, i *struct {}) (*struct {}, error) {
		h.scheduler.ForcePoll(ctx)
		return nil, nil
	})
}