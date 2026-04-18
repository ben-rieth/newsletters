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

type contextKey string
const logKey contextKey = "log"

func WideLogMiddleware(ctx huma.Context, next func(huma.Context)) {
	logMap := NewWideLog()
	ctx = huma.WithValue(ctx, logKey, &logMap)
	
	requestId := uuid.New()

	logMap.AddLogField("requestId", requestId.String())
	logMap.AddLogField("host", ctx.Host())
	logMap.AddLogField("method", ctx.Method())
	logMap.AddLogField("operationId", ctx.Operation().OperationID)
	logMap.AddLogField("path", ctx.Operation().Path)

	startTime := time.Now()

	next(ctx)

	endTime := time.Now()
	logMap.AddLogField("duration", endTime.Sub(startTime))
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

func shouldLog(ctx huma.Context, logMap *WideLog) (bool, slog.Level) {
	if ctx.Status() >= http.StatusInternalServerError {
		return true, slog.LevelError
	}

	if logMap.HasError() {
		return true, slog.LevelError
	}

	d, ok := GetField[time.Duration](logMap, "duration")
	if !ok {
		return true, slog.LevelError
	}

	if d.Seconds() > 2 {
		return true, slog.LevelWarn
	}

	return rand.Float64() < 0.05, slog.LevelInfo
}
