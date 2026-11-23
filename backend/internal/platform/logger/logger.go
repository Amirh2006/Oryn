package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// Setup initializes the global slog default logger with structured JSON output.
func Setup(levelStr string, w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}

	var level slog.Level
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// WithRequestID returns a new context with the given request ID attached.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// FromContext returns a logger with the request ID attached if present in the context.
func FromContext(ctx context.Context) *slog.Logger {
	l := slog.Default()
	if ctx == nil {
		return l
	}
	if reqID, ok := ctx.Value(requestIDKey).(string); ok && reqID != "" {
		return l.With(slog.String("request_id", reqID))
	}
	return l
}
