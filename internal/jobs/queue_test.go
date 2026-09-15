package jobs

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
)

type logCapture struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (c *logCapture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.buf.Write(p)
}

func (c *logCapture) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.buf.String()
}

func (c *logCapture) contains(substring string) bool {
	return strings.Contains(c.String(), substring)
}

func captureLogs(t *testing.T) *logCapture {
	t.Helper()

	capture := &logCapture{}
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(capture, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return capture
}

func startQueue(t *testing.T, workers int) *JobQueue {
	t.Helper()

	q := StartJobQueue(config.Config{JobQueueSize: workers})

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := q.Shutdown(ctx); err != nil {
			t.Errorf("queue did not shut down: %v", err)
		}
	})

	return q
}

func awaitSignal(t *testing.T, signal <-chan struct{}, describe string) {
	t.Helper()

	select {
	case <-signal:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", describe)
	}
}

func TestEnqueueRunsEveryJob(t *testing.T) {
	q := startQueue(t, 4)

	const jobCount = 50
	var ran atomic.Int64

	for range jobCount {
		q.Enqueue(func(context.Context) { ran.Add(1) })
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := q.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if got := ran.Load(); got != jobCount {
		t.Errorf("%d of %d jobs ran", got, jobCount)
	}
}

// Jobs are arbitrary handler closures, so one that panics has to be contained.
// A worker lost to a panic is invisible until the queue has no workers left.
func TestPanickingJobDoesNotKillTheWorker(t *testing.T) {
	q := startQueue(t, 1)

	q.Enqueue(func(context.Context) { panic("job blew up") })

	survived := make(chan struct{})
	q.Enqueue(func(context.Context) { close(survived) })

	awaitSignal(t, survived, "the job queued behind a panicking one")
}

func TestPanickingJobIsLoggedAsAnError(t *testing.T) {
	q := startQueue(t, 1)

	logs := captureLogs(t)

	done := make(chan struct{})
	q.Enqueue(func(context.Context) {
		defer close(done)
		panic("job blew up")
	})

	awaitSignal(t, done, "the panicking job")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := q.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if !logs.contains("Job panicked.") {
		t.Errorf("panic was swallowed without a log line; got %q", logs.String())
	}
}

// An in-flight request can enqueue while the server is draining. Sending on the
// closed channel would panic and take the whole shutdown down with it.
func TestEnqueueAfterShutdownDropsTheJob(t *testing.T) {
	q := startQueue(t, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := q.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	var ran atomic.Bool
	q.Enqueue(func(context.Context) { ran.Store(true) })

	if ran.Load() {
		t.Error("job enqueued after shutdown was run")
	}
}

// Shutdown runs while the server is closing, so a job that has already started
// has to finish rather than be abandoned half-written.
func TestShutdownDrainsInFlightJobs(t *testing.T) {
	q := startQueue(t, 2)

	const jobCount = 6
	var finished atomic.Int64

	for range jobCount {
		q.Enqueue(func(context.Context) {
			time.Sleep(20 * time.Millisecond)
			finished.Add(1)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := q.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if got := finished.Load(); got != jobCount {
		t.Errorf("Shutdown returned with %d of %d jobs finished", got, jobCount)
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	q := startQueue(t, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := q.Shutdown(ctx); err != nil {
		t.Fatalf("first Shutdown: %v", err)
	}

	if err := q.Shutdown(ctx); err != nil {
		t.Errorf("second Shutdown: %v, want nil", err)
	}
}

// The caller's deadline is the whole point of passing a context: a job that hangs
// must not hold the process open past it.
func TestShutdownGivesUpOnTheDeadline(t *testing.T) {
	q := startQueue(t, 1)

	release := make(chan struct{})
	started := make(chan struct{})
	stopped := make(chan struct{})

	q.Enqueue(func(context.Context) {
		defer close(stopped)
		close(started)
		<-release
	})

	awaitSignal(t, started, "the blocking job to start")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := q.Shutdown(ctx)

	close(release)
	awaitSignal(t, stopped, "the blocking job to unblock")

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Shutdown = %v, want context.DeadlineExceeded", err)
	}
}

// Jobs are logged by exception: a kept line is either a failure, a slow job, or
// one of the small sample that shows the queue is working at all.
func TestShouldLogLevels(t *testing.T) {
	tests := []struct {
		name        string
		durationRaw any
		addError    bool
		wantLog     bool
		wantLevel   slog.Level
		wantReason  string
	}{
		{
			name: "failed job", durationRaw: 10 * time.Millisecond, addError: true,
			wantLog: true, wantLevel: slog.LevelError, wantReason: "has-error",
		},
		{
			name: "missing duration", durationRaw: nil,
			wantLog: true, wantLevel: slog.LevelError, wantReason: "missing-duration",
		},
		{
			name: "duration of the wrong type", durationRaw: "1.5s",
			wantLog: true, wantLevel: slog.LevelError, wantReason: "missing-duration",
		},
		{
			name: "slow job", durationRaw: 3 * time.Second,
			wantLog: true, wantLevel: slog.LevelWarn, wantReason: "perf",
		},
		{
			name: "exactly at the slow threshold", durationRaw: 2 * time.Second,
			wantLevel: slog.LevelInfo, wantReason: "chance",
		},
		{
			name: "fast job", durationRaw: 10 * time.Millisecond,
			wantLevel: slog.LevelInfo, wantReason: "chance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wl := wideLog.NewWideLog()
			if tt.durationRaw != nil {
				wl.AddLogField("durationRaw", tt.durationRaw)
			}
			if tt.addError {
				wl.AddErrorField(errors.New("Job panicked."))
			}

			keep, level := shouldLog(wl)

			if level != tt.wantLevel {
				t.Errorf("level = %v, want %v", level, tt.wantLevel)
			}

			if reason, _ := wideLog.GetField[string](wl, "reason"); reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", reason, tt.wantReason)
			}

			// A sampled job is kept at random, so only the forced cases can be
			// asserted.
			if tt.wantLog && !keep {
				t.Error("the job was not logged")
			}
		})
	}
}

func TestShouldLogSamplesOrdinaryJobs(t *testing.T) {
	const jobs = 10000

	kept := 0
	for range jobs {
		wl := wideLog.NewWideLog()
		wl.AddLogField("durationRaw", 10*time.Millisecond)

		if keep, _ := shouldLog(wl); keep {
			kept++
		}
	}

	if kept == 0 || kept > jobs/5 {
		t.Errorf("kept %d of %d ordinary jobs, want a small sample of them", kept, jobs)
	}
}
