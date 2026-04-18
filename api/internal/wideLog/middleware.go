package wideLog

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sort"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type contextKey string
const logKey contextKey = "log"

type logEntry struct {
	Key string
	Value logValue
}

type logValue struct {
	value any
	timestamp time.Time
}

func WideLogMiddleware(ctx huma.Context, next func(huma.Context)) {
	logMap := make(map[string]logValue)
	ctx = huma.WithValue(ctx, logKey, &logMap)
	
	requestId := uuid.New()

	AddLogField(ctx.Context(), "requestId", requestId.String())
	AddLogField(ctx.Context(), "host", ctx.Host())
	AddLogField(ctx.Context(), "method", ctx.Method())
	AddLogField(ctx.Context(), "operationId", ctx.Operation().OperationID)
	AddLogField(ctx.Context(), "path", ctx.Operation().Path)

	startTime := time.Now()

	next(ctx)

	endTime := time.Now()
	AddLogField(ctx.Context(), "duration", endTime.Sub(startTime))
	AddLogField(ctx.Context(), "status", ctx.Status())

	keepLog, logLevel := shouldLog(ctx, logMap)
	if keepLog {
		entries := make([]logEntry, 0, len(logMap))
		for k, v := range logMap {
			entries = append(entries, logEntry{k, v})
		}

		sort.Slice(entries, func(a, b int) bool {
			return entries[a].Value.timestamp.Before(entries[b].Value.timestamp)
		})

		args := make([]any, 0, len(logMap)*2)
		for _, entry := range entries {
			args = append(args, entry.Key, entry.Value.value)
		}
		
		slog.Log(
			ctx.Context(), 
			logLevel, 
			"Request", 
			args...
		)
	}
}

func AddLogField(ctx context.Context, key string, value any) {
	logMap, ok := ctx.Value(logKey).(*map[string]logValue)
	if ok {
		(*logMap)[key] = logValue{
			value: value,
			timestamp: time.Now(),
		}
	}
}

func AddErrorField(ctx context.Context, err error) {
	logMap, ok := ctx.Value(logKey).(*map[string]logValue)
	if ok {
		(*logMap)["error"] = logValue{
			value: err,
			timestamp: time.Now(),
		}
	}
}

func shouldLog(ctx huma.Context, logMap map[string]logValue) (bool, slog.Level) {
	if ctx.Status() >= http.StatusInternalServerError {
		return true, slog.LevelError
	}

	d, ok := logMap["duration"]
	if !ok {
		return true, slog.LevelError
	}

	duration, ok := d.value.(time.Duration)
	if !ok {
		return true, slog.LevelError
	}

	if duration.Seconds() > 2 {
		return true, slog.LevelWarn
	}

	return rand.Float64() < 0.05, slog.LevelInfo
}
