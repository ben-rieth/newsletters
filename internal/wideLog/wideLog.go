package wideLog

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"
)

type contextKey string

const logKey contextKey = "log"

type logEntry struct {
	Key   string
	Value logValue
}

type logValue struct {
	value     any
	timestamp time.Time
}

// A single WideLog is shared by every goroutine working on one request, tick or
// job, so all access to fields must go through mu.
type WideLog struct {
	mu     sync.Mutex
	fields map[string]logValue
}

func NewWideLog() *WideLog {
	return &WideLog{fields: make(map[string]logValue)}
}

func (wl *WideLog) AddLogField(key string, value any) {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	wl.fields[key] = logValue{
		value:     value,
		timestamp: time.Now(),
	}
}

func (wl *WideLog) AddErrorField(errs ...error) {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	errorStrings := make([]string, 0, len(errs))
	for _, err := range errs {
		errorStrings = append(errorStrings, err.Error())
	}

	errLogValue := wl.fields["error"]
	typedErrs, ok := errLogValue.value.([]string)
	if !ok {
		wl.fields["error"] = logValue{
			value:     errorStrings,
			timestamp: time.Now(),
		}
		return
	}

	wl.fields["error"] = logValue{
		value:     append(typedErrs, errorStrings...),
		timestamp: time.Now(),
	}
}

func (wl *WideLog) AddArrayField(key string, value any) {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	arrayLogValue := wl.fields[key]
	typedValue, ok := arrayLogValue.value.([]any)
	if !ok {
		wl.fields[key] = logValue{
			value:     value,
			timestamp: time.Now(),
		}
		return
	}

	wl.fields[key] = logValue{
		value:     append(typedValue, value),
		timestamp: time.Now(),
	}
}

func (wl *WideLog) Slog(ctx context.Context, level slog.Level) {
	wl.mu.Lock()
	entries := make([]logEntry, 0, len(wl.fields))
	for k, v := range wl.fields {
		entries = append(entries, logEntry{k, v})
	}
	wl.mu.Unlock()

	sort.Slice(entries, func(a, b int) bool {
		return entries[a].Value.timestamp.Before(entries[b].Value.timestamp)
	})

	args := make([]any, 0, len(entries)*2)
	for _, entry := range entries {
		args = append(args, entry.Key, entry.Value.value)
	}

	slog.Log(ctx, level, "Request", args...)
}

func (wl *WideLog) HasError() bool {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	_, ok := wl.fields["error"]
	return ok
}

func GetField[T any](wl *WideLog, key string) (T, bool) {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	val, ok := wl.fields[key]
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := val.value.(T)
	return typed, ok
}

func CreateWideLogAndAddToContext(baseCtx context.Context) (context.Context, *WideLog) {
	wl := NewWideLog()
	return context.WithValue(baseCtx, logKey, wl), wl
}
