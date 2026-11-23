package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/response"
	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from any panics and logs them structured.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID := c.GetString(RequestIDKey)
				stack := string(debug.Stack())

				slog.Error("Unhandled Panic Recovered",
					slog.String("request_id", reqID),
					slog.Any("panic", r),
					slog.String("stack", stack),
				)

				response.Error(
					c,
					http.StatusInternalServerError,
					domain.ErrCodeInternal,
					"an unexpected internal server error occurred",
					nil,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}
