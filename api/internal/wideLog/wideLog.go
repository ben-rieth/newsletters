package wideLog

import (
	"context"
	"log/slog"
	"sort"
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

type WideLog map[string]logValue

func NewWideLog() *WideLog {
	return &WideLog{}
}

func (wl *WideLog) AddLogField(key string, value any) {
	(*wl)[key] = logValue{
		value:     value,
		timestamp: time.Now(),
	}
}

func (wl *WideLog) AddErrorField(errs ...error) {
	errLogValue, ok := (*wl)["error"]
	typedErrs, ok := errLogValue.value.([]string)

	errorStrings := make([]string, 0, len(errs))
	for _, err := range errs {
		errorStrings = append(errorStrings, err.Error())
	}

	if !ok {
		(*wl)["error"] = logValue{
			value:     errorStrings,
			timestamp: time.Now(),
		}
		return
	}

	newErrs := append(typedErrs, errorStrings...)
	(*wl)["error"] = logValue{
		value:     newErrs,
		timestamp: time.Now(),
	}
}

func (wl *WideLog) AddArrayField(key string, value any) {
	arrayLogValue, ok := (*wl)[key]
	typedValue, ok := arrayLogValue.value.([]any)
	if !ok {
		(*wl)[key] = logValue{
			value:     value,
			timestamp: time.Now(),
		}
		return
	}

	newValues := append(typedValue, value)
	(*wl)[key] = logValue{
		value:     newValues,
		timestamp: time.Now(),
	}
}

func (wl *WideLog) Slog(ctx context.Context, level slog.Level) {
	entries := make([]logEntry, 0, len(*wl))
	for k, v := range *wl {
		entries = append(entries, logEntry{k, v})
	}

	sort.Slice(entries, func(a, b int) bool {
		return entries[a].Value.timestamp.Before(entries[b].Value.timestamp)
	})

	args := make([]any, 0, len(*wl)*2)
	for _, entry := range entries {
		args = append(args, entry.Key, entry.Value.value)
	}

	slog.Log(ctx, level, "Request", args...)
}

func (wl *WideLog) HasError() bool {
	_, ok := (*wl)["error"]
	return ok
}

func GetField[T any](wl *WideLog, key string) (T, bool) {
	val, ok := (*wl)[key]
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
