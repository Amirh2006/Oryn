package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogger logs HTTP requests using Go's log/slog.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		reqID := c.GetString(RequestIDKey)
		clientIP := c.ClientIP()
		method := c.Request.Method

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		attrs := []slog.Attr{
			slog.String("request_id", reqID),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", clientIP),
			slog.Int("bytes_written", c.Writer.Size()),
		}

		lvl := slog.LevelInfo
		if status >= 500 {
			lvl = slog.LevelError
		} else if status >= 400 {
			lvl = slog.LevelWarn
		}

		slog.LogAttrs(c.Request.Context(), lvl, "HTTP Request", attrs...)
	}
}
