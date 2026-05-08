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
type JobQueue chan Job

func StartJobQueue(ctx context.Context, cfg config.Config) JobQueue {
	jobChannel := make(chan Job, cfg.JobQueueSize*10)

	wg := sync.WaitGroup{}

	for range cfg.JobQueueSize {
		wg.Go(func() {
			for job := range jobChannel {
				jobCtx := context.Background()
				jobCtx, wl := wideLog.CreateWideLogAndAddToContext(jobCtx)

				wl.AddLogField("job_id", uuid.New())
				startTime := time.Now()

				runJob(jobCtx, job)

				endTime := time.Now()
				wl.AddLogField("duration", endTime.Sub(startTime).String())

				shouldLog, level := shouldLog(wl)
				if shouldLog {
					wl.Slog(jobCtx, level)
				}
			}
		})
	}

	go func() {
		<-ctx.Done()
		close(jobChannel)
		log.Printf("Waiting for final jobs to complete before shutting down job queue")
		wg.Wait()
		log.Println("Jobs complete. Shutting down job queue")

	}()

	return jobChannel
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

	d, ok := wideLog.GetField[time.Duration](logMap, "duration")
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
