package wideLog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

type capturedLog struct {
	level   slog.Level
	message string
	keys    []string
	values  map[string]any
}

type capturingHandler struct {
	mu   sync.Mutex
	logs []capturedLog
}

func (h *capturingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *capturingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *capturingHandler) WithGroup(string) slog.Handler { return h }

func (h *capturingHandler) Handle(_ context.Context, record slog.Record) error {
	captured := capturedLog{
		level:   record.Level,
		message: record.Message,
		values:  make(map[string]any, record.NumAttrs()),
	}

	record.Attrs(func(attr slog.Attr) bool {
		captured.keys = append(captured.keys, attr.Key)
		captured.values[attr.Key] = attr.Value.Any()
		return true
	})

	h.mu.Lock()
	defer h.mu.Unlock()

	h.logs = append(h.logs, captured)

	return nil
}

func (h *capturingHandler) captured() []capturedLog {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]capturedLog(nil), h.logs...)
}

func captureSlog(t *testing.T) *capturingHandler {
	t.Helper()

	handler := &capturingHandler{}
	previous := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return handler
}

func onlyLog(t *testing.T, handler *capturingHandler) capturedLog {
	t.Helper()

	logs := handler.captured()
	if len(logs) != 1 {
		t.Fatalf("emitted %d log lines, want 1", len(logs))
	}

	return logs[0]
}

func TestAddLogFieldOverwrites(t *testing.T) {
	wl := NewWideLog()

	wl.AddLogField("status", 200)
	wl.AddLogField("status", 500)

	got, ok := GetField[int](wl, "status")
	if !ok {
		t.Fatal("status is missing after being set twice")
	}

	if got != 500 {
		t.Errorf("status = %d, want the value set last", got)
	}
}

// shouldLog reads the request duration back out through GetField and treats a
// miss as a reason to log at error level, so both miss shapes have to be a miss
// rather than a usable zero.
func TestGetFieldMisses(t *testing.T) {
	wl := NewWideLog()
	wl.AddLogField("duration", "1.5s")

	t.Run("absent key", func(t *testing.T) {
		got, ok := GetField[time.Duration](wl, "nothing-here")
		if ok {
			t.Errorf("GetField reported a hit for an absent key, returning %v", got)
		}

		if got != 0 {
			t.Errorf("GetField returned %v for an absent key, want the zero value", got)
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		got, ok := GetField[time.Duration](wl, "duration")
		if ok {
			t.Errorf("GetField reported a hit for a string stored under duration, returning %v", got)
		}

		if got != 0 {
			t.Errorf("GetField returned %v for a type mismatch, want the zero value", got)
		}
	})
}

func TestAddMessageAppends(t *testing.T) {
	wl := NewWideLog()

	wl.AddMessage("first")
	wl.AddMessage("second")

	got, ok := GetField[[]any](wl, "messages")
	if !ok {
		t.Fatal("messages is missing")
	}

	if want := []any{"first", "second"}; !slices.Equal(got, want) {
		t.Errorf("messages = %v, want %v", got, want)
	}
}

func TestAddArrayFieldAppendsPerKey(t *testing.T) {
	wl := NewWideLog()

	wl.AddArrayField("sendEmailAttempts", "attempt 1 failed")
	wl.AddArrayField("sendEmailAttempts", "attempt 2 failed")
	wl.AddArrayField("GetFeedDataSince", "feed a")

	attempts, ok := GetField[[]any](wl, "sendEmailAttempts")
	if !ok {
		t.Fatal("sendEmailAttempts is missing")
	}

	if want := []any{"attempt 1 failed", "attempt 2 failed"}; !slices.Equal(attempts, want) {
		t.Errorf("sendEmailAttempts = %v, want %v", attempts, want)
	}

	fetches, ok := GetField[[]any](wl, "GetFeedDataSince")
	if !ok {
		t.Fatal("GetFeedDataSince is missing")
	}

	if len(fetches) != 1 {
		t.Errorf("GetFeedDataSince = %v, want only its own entry", fetches)
	}
}

// Errors arrive from several goroutines working on one request, so each one has
// to be kept rather than replacing the last.
func TestAddErrorFieldAccumulates(t *testing.T) {
	wl := NewWideLog()

	wl.AddErrorField(errors.New("first"), errors.New("second"))
	wl.AddErrorField(errors.New("third"))

	got, ok := GetField[[]string](wl, "error")
	if !ok {
		t.Fatal("error is missing")
	}

	if want := []string{"first", "second", "third"}; !slices.Equal(got, want) {
		t.Errorf("error = %v, want %v", got, want)
	}
}

// A field named "error" that was not written by AddErrorField would otherwise
// make the append panic on a failed type assertion, in the one code path that
// only ever runs when something has already gone wrong.
func TestAddErrorFieldSurvivesACollidingField(t *testing.T) {
	wl := NewWideLog()

	wl.AddLogField("error", 42)
	wl.AddErrorField(errors.New("the real error"))

	got, ok := GetField[[]string](wl, "error")
	if !ok {
		t.Fatalf("error is no longer a %T after colliding with another field", []string{})
	}

	if want := []string{"the real error"}; !slices.Equal(got, want) {
		t.Errorf("error = %v, want %v", got, want)
	}
}

// Logging must not be the thing that takes down the caller, and the only paths
// that add errors are the ones already handling a failure.
func TestAddErrorFieldIgnoresNilErrors(t *testing.T) {
	wl := NewWideLog()

	wl.AddErrorField(nil)
	wl.AddErrorField(errors.New("real failure"), nil, errors.New("another failure"))

	got, ok := GetField[[]string](wl, "error")
	if !ok {
		t.Fatal("error is missing")
	}

	if want := []string{"real failure", "another failure"}; !slices.Equal(got, want) {
		t.Errorf("error = %v, want %v", got, want)
	}
}

// Reporting no errors is not the same as reporting a failure: a log marked with
// an error is escalated to error level and always kept.
func TestAddErrorFieldWithNothingToReportLeavesTheLogClean(t *testing.T) {
	tests := []struct {
		name string
		errs []error
	}{
		{name: "no arguments", errs: nil},
		{name: "one nil error", errs: []error{nil}},
		{name: "only nil errors", errs: []error{nil, nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wl := NewWideLog()
			wl.AddLogField("path", "/newsletters")

			wl.AddErrorField(tt.errs...)

			if wl.HasError() {
				got, _ := GetField[[]string](wl, "error")
				t.Errorf("the log reports an error after adding none; error = %v", got)
			}
		})
	}
}

// An empty variadic call must not wipe a field that another writer put there.
func TestAddErrorFieldWithNothingToReportKeepsEarlierErrors(t *testing.T) {
	wl := NewWideLog()

	wl.AddErrorField(errors.New("real failure"))
	wl.AddErrorField()

	got, ok := GetField[[]string](wl, "error")
	if !ok {
		t.Fatal("error is missing")
	}

	if want := []string{"real failure"}; !slices.Equal(got, want) {
		t.Errorf("error = %v, want %v", got, want)
	}
}

func TestHasError(t *testing.T) {
	wl := NewWideLog()

	if wl.HasError() {
		t.Error("a fresh log reports an error")
	}

	wl.AddLogField("status", 500)
	if wl.HasError() {
		t.Error("an unrelated field was read as an error")
	}

	wl.AddErrorField(errors.New("boom"))
	if !wl.HasError() {
		t.Error("an added error is not reported")
	}
}

// Fields are ordered by when they were added so that a log line reads in the
// order the request actually happened.
func TestSlogAsEmitsFieldsInInsertionOrder(t *testing.T) {
	handler := captureSlog(t)

	wl := NewWideLog()
	want := []string{"requestId", "method", "path", "status"}
	for _, key := range want {
		wl.AddLogField(key, key)
		time.Sleep(time.Millisecond)
	}

	wl.SlogAs(context.Background(), slog.LevelWarn, "Scheduler Tick")

	got := onlyLog(t, handler)

	if !slices.Equal(got.keys, want) {
		t.Errorf("field order = %v, want %v", got.keys, want)
	}

	if got.level != slog.LevelWarn {
		t.Errorf("level = %v, want warn", got.level)
	}

	if got.message != "Scheduler Tick" {
		t.Errorf("message = %q, want the one passed to SlogAs", got.message)
	}
}

// Overwriting a field refreshes its timestamp, so the final value is reported
// where it was set rather than where the key first appeared.
func TestSlogAsOrdersAnOverwrittenFieldByItsLastWrite(t *testing.T) {
	handler := captureSlog(t)

	wl := NewWideLog()
	wl.AddLogField("status", 200)
	time.Sleep(time.Millisecond)
	wl.AddLogField("path", "/newsletters")
	time.Sleep(time.Millisecond)
	wl.AddLogField("status", 500)

	wl.SlogAs(context.Background(), slog.LevelError, "Request")

	got := onlyLog(t, handler)

	if want := []string{"path", "status"}; !slices.Equal(got.keys, want) {
		t.Errorf("field order = %v, want %v", got.keys, want)
	}

	if got.values["status"] != int64(500) {
		t.Errorf("status = %v, want the value set last", got.values["status"])
	}
}

func TestSlogUsesTheRequestMessage(t *testing.T) {
	handler := captureSlog(t)

	wl := NewWideLog()
	wl.AddLogField("path", "/newsletters")

	wl.Slog(context.Background(), slog.LevelInfo)

	if got := onlyLog(t, handler); got.message != "Request" {
		t.Errorf("message = %q, want Request", got.message)
	}
}

func TestSlogAsEmitsAnEmptyLogWithoutFields(t *testing.T) {
	handler := captureSlog(t)

	NewWideLog().SlogAs(context.Background(), slog.LevelInfo, "Request")

	if got := onlyLog(t, handler); len(got.keys) != 0 {
		t.Errorf("an empty log emitted %v", got.keys)
	}
}

// One log is shared by every goroutine working on a request, tick or job, so a
// lost write means a missing field in the only record of what happened.
func TestConcurrentWritersKeepEveryField(t *testing.T) {
	const writers = 50

	wl := NewWideLog()

	var wg sync.WaitGroup
	for i := range writers {
		wg.Go(func() {
			wl.AddLogField(fmt.Sprintf("field-%d", i), i)
			wl.AddArrayField("shared", i)
			wl.AddMessage("worked")
			wl.AddErrorField(fmt.Errorf("failure %d", i))
		})
	}
	wg.Wait()

	for i := range writers {
		key := fmt.Sprintf("field-%d", i)
		if got, ok := GetField[int](wl, key); !ok || got != i {
			t.Errorf("%s = %v (present: %v), want %d", key, got, ok, i)
		}
	}

	shared, ok := GetField[[]any](wl, "shared")
	if !ok || len(shared) != writers {
		t.Errorf("shared has %d entries, want %d", len(shared), writers)
	}

	messages, ok := GetField[[]any](wl, "messages")
	if !ok || len(messages) != writers {
		t.Errorf("messages has %d entries, want %d", len(messages), writers)
	}

	errs, ok := GetField[[]string](wl, "error")
	if !ok || len(errs) != writers {
		t.Errorf("error has %d entries, want %d", len(errs), writers)
	}
}

// The feed fan-out logs while its goroutines are still writing, so emitting must
// not race with them or observe a half-written map.
func TestSlogAsIsSafeWhileFieldsAreBeingAdded(t *testing.T) {
	captureSlog(t)

	wl := NewWideLog()

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			wl.AddLogField(fmt.Sprintf("field-%d", i), i)
			wl.AddArrayField("shared", i)
		})
		wg.Go(func() {
			wl.SlogAs(context.Background(), slog.LevelInfo, "Request")
		})
	}
	wg.Wait()
}

func TestCreateWideLogAndAddToContext(t *testing.T) {
	ctx, wl := CreateWideLogAndAddToContext(context.Background())

	AddLogField(ctx, "path", "/newsletters")
	AddArrayField(ctx, "sendEmailAttempts", "attempt 1 failed")
	AddMessage(ctx, "worked")
	AddErrorField(ctx, errors.New("boom"))

	if got, ok := GetField[string](wl, "path"); !ok || got != "/newsletters" {
		t.Errorf("path = %q (present: %v), want the value added through the context", got, ok)
	}

	if got, ok := GetField[[]any](wl, "sendEmailAttempts"); !ok || len(got) != 1 {
		t.Errorf("sendEmailAttempts = %v, want the entry added through the context", got)
	}

	if got, ok := GetField[[]any](wl, "messages"); !ok || len(got) != 1 {
		t.Errorf("messages = %v, want the message added through the context", got)
	}

	if !wl.HasError() {
		t.Error("the error added through the context is missing")
	}
}

// Background work and tests call these with whatever context they have. Logging
// is never important enough to take down the caller that was doing real work.
func TestContextHelpersAreNoOpsWithoutALog(t *testing.T) {
	for _, tt := range []struct {
		name string
		call func(context.Context)
	}{
		{name: "AddLogField", call: func(ctx context.Context) { AddLogField(ctx, "path", "/newsletters") }},
		{name: "AddArrayField", call: func(ctx context.Context) { AddArrayField(ctx, "attempts", "failed") }},
		{name: "AddMessage", call: func(ctx context.Context) { AddMessage(ctx, "worked") }},
		{name: "AddErrorField", call: func(ctx context.Context) { AddErrorField(ctx, errors.New("boom")) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.call(context.Background())
			tt.call(context.WithValue(context.Background(), logKey, "not a log"))
		})
	}
}

func requestContext(t *testing.T, status int) huma.Context {
	t.Helper()

	ctx := humatest.NewContext(
		&huma.Operation{OperationID: "list-newsletters", Path: "/newsletters"},
		httptest.NewRequest(http.MethodGet, "/newsletters", nil),
		httptest.NewRecorder(),
	)
	ctx.SetStatus(status)

	return ctx
}

// Every request's only record of what happened is this one line, so the shape of
// it is load-bearing for anything read out of the logs later.
func TestWideLogMiddlewareLogsTheRequestShape(t *testing.T) {
	handler := captureSlog(t)

	WideLogMiddleware(requestContext(t, http.StatusOK), func(ctx huma.Context) {
		ctx.SetStatus(http.StatusInternalServerError)
	})

	got := onlyLog(t, handler)

	if got.level != slog.LevelError {
		t.Errorf("level = %v, want error for a 500", got.level)
	}

	for _, key := range []string{"requestId", "host", "method", "operationId", "path", "start", "end", "durationStr", "duration", "status", "reason"} {
		if _, ok := got.values[key]; !ok {
			t.Errorf("%s is missing from the request log; got %v", key, got.keys)
		}
	}

	wantValues := map[string]any{
		"method":      http.MethodGet,
		"path":        "/newsletters",
		"operationId": "list-newsletters",
		"status":      int64(http.StatusInternalServerError),
		"reason":      "500",
	}
	for key, want := range wantValues {
		if got.values[key] != want {
			t.Errorf("%s = %v, want %v", key, got.values[key], want)
		}
	}
}

func TestWideLogMiddlewareKeepsWhatTheHandlerAdded(t *testing.T) {
	handler := captureSlog(t)

	WideLogMiddleware(requestContext(t, http.StatusOK), func(ctx huma.Context) {
		AddLogField(ctx.Context(), "newsletterId", "nl-1")
		AddMessage(ctx.Context(), "worked")
		ctx.SetStatus(http.StatusInternalServerError)
	})

	got := onlyLog(t, handler)

	if got.values["newsletterId"] != "nl-1" {
		t.Errorf("newsletterId = %v, want the value the handler added", got.values["newsletterId"])
	}

	if _, ok := got.values["messages"]; !ok {
		t.Errorf("the handler's message is missing; got %v", got.keys)
	}
}

// A handler that reports an error while still returning a 2xx is the case worth
// getting right: nothing else in the response would reveal it.
func TestWideLogMiddlewareLogsAHandlerErrorOnASuccessfulResponse(t *testing.T) {
	handler := captureSlog(t)

	WideLogMiddleware(requestContext(t, http.StatusOK), func(ctx huma.Context) {
		AddErrorField(ctx.Context(), errors.New("failed to mark item as read"))
		ctx.SetStatus(http.StatusOK)
	})

	got := onlyLog(t, handler)

	if got.level != slog.LevelError {
		t.Errorf("level = %v, want error", got.level)
	}

	if got.values["reason"] != "has-error" {
		t.Errorf("reason = %v, want has-error", got.values["reason"])
	}
}

func TestShouldLogLevels(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		duration   any
		addError   bool
		wantLog    bool
		wantLevel  slog.Level
		wantReason string
	}{
		{
			name: "server error", status: http.StatusInternalServerError, duration: 10 * time.Millisecond,
			wantLog: true, wantLevel: slog.LevelError, wantReason: "500",
		},
		{
			name: "gateway error", status: http.StatusBadGateway, duration: 10 * time.Millisecond,
			wantLog: true, wantLevel: slog.LevelError, wantReason: "500",
		},
		// A 4xx is the caller's problem, not ours, so it is sampled like any other
		// answered request.
		{
			name: "client error", status: http.StatusBadRequest, duration: 10 * time.Millisecond,
			wantLevel: slog.LevelInfo, wantReason: "chance",
		},
		{
			name: "reported error behind a 2xx", status: http.StatusOK, duration: 10 * time.Millisecond, addError: true,
			wantLog: true, wantLevel: slog.LevelError, wantReason: "has-error",
		},
		// Without a duration the request cannot be judged at all, which is itself
		// worth an error line.
		{
			name: "missing duration", status: http.StatusOK, duration: nil,
			wantLog: true, wantLevel: slog.LevelError, wantReason: "missing-duration",
		},
		{
			name: "duration of the wrong type", status: http.StatusOK, duration: "1.5s",
			wantLog: true, wantLevel: slog.LevelError, wantReason: "missing-duration",
		},
		{
			name: "slow request", status: http.StatusOK, duration: 3 * time.Second,
			wantLog: true, wantLevel: slog.LevelWarn, wantReason: "perf",
		},
		{
			name: "exactly at the slow threshold", status: http.StatusOK, duration: 2 * time.Second,
			wantLevel: slog.LevelInfo, wantReason: "chance",
		},
		{
			name: "fast request", status: http.StatusOK, duration: 10 * time.Millisecond,
			wantLevel: slog.LevelInfo, wantReason: "chance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wl := NewWideLog()
			if tt.duration != nil {
				wl.AddLogField("duration", tt.duration)
			}
			if tt.addError {
				wl.AddErrorField(errors.New("boom"))
			}

			keep, level := shouldLog(requestContext(t, tt.status), wl)

			if level != tt.wantLevel {
				t.Errorf("level = %v, want %v", level, tt.wantLevel)
			}

			if reason, _ := GetField[string](wl, "reason"); reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", reason, tt.wantReason)
			}

			// A sampled request is kept at random, so only the forced cases can be
			// asserted.
			if tt.wantLog && !keep {
				t.Error("the request was not logged")
			}
		})
	}
}

// The sampling rate is a tuning knob, but always-on would flood the logs and
// never-on would leave successful requests invisible.
func TestShouldLogSamplesOrdinaryRequests(t *testing.T) {
	const requests = 10000

	ctx := requestContext(t, http.StatusOK)

	kept := 0
	for range requests {
		wl := NewWideLog()
		wl.AddLogField("duration", 10*time.Millisecond)

		if keep, _ := shouldLog(ctx, wl); keep {
			kept++
		}
	}

	if kept == 0 || kept > requests/5 {
		t.Errorf("kept %d of %d ordinary requests, want a small sample of them", kept, requests)
	}
}
