package wideLog

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

func WideLogMiddleware(ctx huma.Context, next func(huma.Context)) {
	logMap := NewWideLog()
	ctx = huma.WithValue(ctx, logKey, logMap)

	requestId := uuid.New()

	logMap.AddLogField("requestId", requestId.String())
	logMap.AddLogField("host", ctx.Host())
	logMap.AddLogField("method", ctx.Method())
	logMap.AddLogField("operationId", ctx.Operation().OperationID)
	logMap.AddLogField("path", ctx.Operation().Path)

	startTime := time.Now()

	next(ctx)

	endTime := time.Now()
	logMap.AddLogField("duration", endTime.Sub(startTime).String())
	logMap.AddLogField("status", ctx.Status())

	keepLog, logLevel := shouldLog(ctx, logMap)
	if keepLog {
		logMap.Slog(ctx.Context(), logLevel)
	}
}

func AddLogField(ctx context.Context, key string, value any) {
	wl, ok := ctx.Value(logKey).(*WideLog)
	if ok {
		wl.AddLogField(key, value)
	}
}

func AddErrorField(ctx context.Context, err ...error) {
	wl, ok := ctx.Value(logKey).(*WideLog)
	if ok {
		wl.AddErrorField(err...)
	}
}

func AddArrayField(ctx context.Context, key string, value any) {
	wl, ok := ctx.Value(logKey).(*WideLog)
	if ok {
		wl.AddArrayField(key, value)
	}
}

func shouldLog(ctx huma.Context, logMap *WideLog) (bool, slog.Level) {
	if ctx.Status() >= http.StatusInternalServerError {
		logMap.AddLogField("reason", "500")
		return true, slog.LevelError
	}

	if logMap.HasError() {
		logMap.AddLogField("reason", "has-error")
		return true, slog.LevelError
	}

	d, ok := GetField[time.Duration](logMap, "duration")
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
