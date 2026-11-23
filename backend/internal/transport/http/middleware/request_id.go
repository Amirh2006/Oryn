package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the HTTP header used for Request ID propagation
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the Gin context key storing the Request ID
	RequestIDKey = "RequestID"
)

// RequestID returns a Gin middleware that extracts an existing X-Request-ID or generates a new UUID v4.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.New().String()
		}

		c.Set(RequestIDKey, reqID)
		c.Header(RequestIDHeader, reqID)
		c.Next()
	}
}
