package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/platform/security"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/response"
	"github.com/gin-gonic/gin"
)

const (
	UserIDKey    = "UserID"
	UserEmailKey = "UserEmail"
	AccessTokenCookie = "oryn_access_token"
	RefreshTokenCookie = "oryn_refresh_token"
)

type userContextKey string
const ctxUserIDKey userContextKey = "user_id"

// Auth enforces authentication via HttpOnly cookie or Authorization Bearer header
func Auth(jwtManager *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, domain.ErrCodeUnauthorized, "missing authentication token", nil)
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenStr)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, domain.ErrCodeUnauthorized, "invalid or expired token", nil)
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)

		// Also attach to request context for clean downstream use
		ctx := context.WithValue(c.Request.Context(), ctxUserIDKey, claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalAuth extracts authentication if present without blocking guests
func OptionalAuth(jwtManager *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr != "" {
			if claims, err := jwtManager.ValidateToken(tokenStr); err == nil {
				c.Set(UserIDKey, claims.UserID)
				c.Set(UserEmailKey, claims.Email)
				ctx := context.WithValue(c.Request.Context(), ctxUserIDKey, claims.UserID)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// 1. Check HttpOnly cookie first
	if cookie, err := c.Cookie(AccessTokenCookie); err == nil && cookie != "" {
		return cookie
	}

	// 2. Check Authorization Bearer header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	return ""
}
