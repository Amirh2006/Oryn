package handlers

import (
	"net/http"

	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/platform/security"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/middleware"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/response"
	"github.com/Amirh2006/Oryn/backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	usecase    *usecase.AuthUsecase
	authConfig config.AuthConfig
}

func NewAuthHandler(u *usecase.AuthUsecase, cfg config.AuthConfig) *AuthHandler {
	return &AuthHandler{
		usecase:    u,
		authConfig: cfg,
	}
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"displayName"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid request body", err.Error())
		return
	}

	user, tokens, err := h.usecase.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	h.setSessionCookies(c, tokens)

	response.JSON(c, http.StatusCreated, gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid request body", err.Error())
		return
	}

	user, tokens, err := h.usecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	h.setSessionCookies(c, tokens)

	response.JSON(c, http.StatusOK, gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	// Try reading refresh token from cookie first, fallback to JSON body
	refreshToken, err := c.Cookie(middleware.RefreshTokenCookie)
	if err != nil || refreshToken == "" {
		var req RefreshRequest
		_ = c.ShouldBindJSON(&req)
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "missing refresh token", nil)
		return
	}

	newTokens, err := h.usecase.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.clearSessionCookies(c)
		response.HandleError(c, err)
		return
	}

	h.setSessionCookies(c, newTokens)

	response.JSON(c, http.StatusOK, gin.H{
		"tokens": newTokens,
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(middleware.RefreshTokenCookie)
	if refreshToken == "" {
		var req RefreshRequest
		_ = c.ShouldBindJSON(&req)
		refreshToken = req.RefreshToken
	}

	_ = h.usecase.Logout(c.Request.Context(), refreshToken)
	h.clearSessionCookies(c)

	response.JSON(c, http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	user, err := h.usecase.GetMe(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, user)
}

func (h *AuthHandler) setSessionCookies(c *gin.Context, tokens *security.TokenPair) {
	accessMaxAge := int(tokens.AccessExpiry.Seconds())
	refreshMaxAge := int(tokens.RefreshExpiry.Seconds())

	domain := h.authConfig.CookieDomain
	secure := h.authConfig.CookieSecure

	// Access Token HttpOnly cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.AccessTokenCookie, tokens.AccessToken, accessMaxAge, "/", domain, secure, true)

	// Refresh Token HttpOnly cookie
	c.SetCookie(middleware.RefreshTokenCookie, tokens.RefreshToken, refreshMaxAge, "/api/v1/auth", domain, secure, true)
}

func (h *AuthHandler) clearSessionCookies(c *gin.Context) {
	domain := h.authConfig.CookieDomain
	secure := h.authConfig.CookieSecure

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.AccessTokenCookie, "", -1, "/", domain, secure, true)
	c.SetCookie(middleware.RefreshTokenCookie, "", -1, "/api/v1/auth", domain, secure, true)
}
