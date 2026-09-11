package jobs

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/google/uuid"
)

type Job func(context.Context)

type JobQueue struct {
	jobs   chan Job
	wg     sync.WaitGroup
	mu     sync.RWMutex
	closed bool
}

func StartJobQueue(cfg config.Config) *JobQueue {
	q := &JobQueue{
		jobs: make(chan Job, cfg.JobQueueSize*10),
	}

	for range cfg.JobQueueSize {
		q.wg.Go(func() {
			for job := range q.jobs {
				jobCtx := context.Background()
				jobCtx, wl := wideLog.CreateWideLogAndAddToContext(jobCtx)

				wl.AddLogField("job_id", uuid.New())
				startTime := time.Now()

				runJob(jobCtx, job)

				endTime := time.Now()

				duration := endTime.Sub(startTime)

				wl.AddLogField("durationRaw", duration)
				wl.AddLogField("duration", duration.String())

				shouldLog, level := shouldLog(wl)
				if shouldLog {
					wl.SlogAs(jobCtx, level, "Job")
				}
			}
		})
	}

	return q
}

// Enqueue drops the job once the queue is shutting down: sending on the closed
// channel would panic, and an in-flight request enqueueing during drain is
// normal rather than exceptional.
func (q *JobQueue) Enqueue(job Job) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		log.Println("Job rejected: job queue is shutting down")
		return
	}

	q.jobs <- job
}

// The lock is released before the drain wait so that jobs still enqueueing are
// rejected rather than blocked behind it.
func (q *JobQueue) closeChannel() (closedByCaller bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return false
	}

	q.closed = true
	close(q.jobs)
	return true
}

func (q *JobQueue) Shutdown(ctx context.Context) error {
	if !q.closeChannel() {
		return nil
	}

	log.Println("Waiting for final jobs to complete before shutting down job queue")

	drained := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(drained)
	}()

	select {
	case <-drained:
		log.Println("Jobs complete. Shutting down job queue")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func runJob(jobCtx context.Context, job Job) {
	defer func() {
		if r := recover(); r != nil {
			wideLog.AddErrorField(jobCtx, errors.New("Job panicked."))
		}
	}()

	job(jobCtx)
}

func shouldLog(logMap *wideLog.WideLog) (bool, slog.Level) {
	if logMap.HasError() {
		logMap.AddLogField("reason", "has-error")
		return true, slog.LevelError
	}

	d, ok := wideLog.GetField[time.Duration](logMap, "durationRaw")
	if !ok {
		logMap.AddLogField("reason", "missing-duration")
		return true, slog.LevelError
	}

	if d.Seconds() > 2 {
		logMap.AddLogField("reason", "perf")
		return true, slog.LevelWarn
	}

	logMap.AddLogField("reason", "chance")
	return rand.Float64() < 0.05, slog.LevelInfo
}
