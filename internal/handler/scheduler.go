package handler

import (
	"context"
	"net/http"

	"github.com/ben-rieth/newsletter-api/internal/jobs"
	"github.com/ben-rieth/newsletter-api/internal/newsletters"
	"github.com/danielgtaylor/huma/v2"
)

type SchedulerHandler struct {
	scheduler *newsletters.Scheduler
	jobQueue  jobs.JobQueue
}

func NewSchedulerHandler(scheduler *newsletters.Scheduler, jobQueue jobs.JobQueue) *SchedulerHandler {
	return &SchedulerHandler{
		scheduler,
		jobQueue,
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
		h.jobQueue <- func(ctx context.Context) {
			h.scheduler.ForcePoll()
		}
		return nil, nil
	})
}
