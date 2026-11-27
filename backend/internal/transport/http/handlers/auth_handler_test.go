package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/platform/security"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/middleware"
	"github.com/Amirh2006/Oryn/backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	m.users[u.Email] = u
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	m.users[u.Email] = u
	return nil
}

type mockTokenRepo struct {
	tokens map[string]*domain.RefreshToken
}

func (m *mockTokenRepo) Save(ctx context.Context, t *domain.RefreshToken) error {
	m.tokens[t.TokenHash] = t
	return nil
}
func (m *mockTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	if t, ok := m.tokens[hash]; ok {
		return t, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockTokenRepo) RevokeByHash(ctx context.Context, hash string) error {
	if t, ok := m.tokens[hash]; ok {
		t.Revoked = true
	}
	return nil
}
func (m *mockTokenRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
		}
	}
	return nil
}

type mockPrefRepo struct {
	prefs map[string]*domain.UserPreferences
}

func (m *mockPrefRepo) GetByUserID(ctx context.Context, userID string) (*domain.UserPreferences, error) {
	if p, ok := m.prefs[userID]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockPrefRepo) Upsert(ctx context.Context, p *domain.UserPreferences) error {
	m.prefs[p.UserID] = p
	return nil
}

func setupAuthTestRouter() (*gin.Engine, *AuthHandler, *security.JWTManager) {
	gin.SetMode(gin.TestMode)
	uRepo := &mockUserRepo{users: make(map[string]*domain.User)}
	tRepo := &mockTokenRepo{tokens: make(map[string]*domain.RefreshToken)}
	pRepo := &mockPrefRepo{prefs: make(map[string]*domain.UserPreferences)}

	jwtSecret := "super-secure-test-jwt-secret-key-32-bytes"
	jwtMgr := security.NewJWTManager(jwtSecret, 15*time.Minute, 24*time.Hour)

	authUsecase := usecase.NewAuthUsecase(uRepo, tRepo, pRepo, jwtMgr)
	authCfg := config.AuthConfig{
		JWTSecret:        jwtSecret,
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
		CookieDomain:     "",
		CookieSecure:     false,
	}

	handler := NewAuthHandler(authUsecase, authCfg)

	r := gin.New()
	api := r.Group("/api/v1/auth")
	{
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)
		api.POST("/refresh", handler.Refresh)
		api.POST("/logout", handler.Logout)
		api.GET("/me", middleware.Auth(jwtMgr), handler.Me)
	}

	return r, handler, jwtMgr
}

func TestAuthHandler_RegisterAndLogin(t *testing.T) {
	router, _, _ := setupAuthTestRouter()

	// 1. Register valid user
	regPayload := map[string]string{
		"email":       "test@oryn.app",
		"password":    "Password123!",
		"displayName": "Oryn Tester",
	}
	body, _ := json.Marshal(regPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify cookies were set
	cookies := rec.Result().Cookies()
	var hasAccess, hasRefresh bool
	for _, c := range cookies {
		if c.Name == middleware.AccessTokenCookie && c.Value != "" {
			hasAccess = true
		}
		if c.Name == middleware.RefreshTokenCookie && c.Value != "" {
			hasRefresh = true
		}
	}
	assert.True(t, hasAccess)
	assert.True(t, hasRefresh)

	// 2. Register with invalid password
	invalidPayload := map[string]string{
		"email":    "test2@oryn.app",
		"password": "short",
	}
	body, _ = json.Marshal(invalidPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// 3. Login with correct password
	loginPayload := map[string]string{
		"email":    "test@oryn.app",
		"password": "Password123!",
	}
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// 4. Login with wrong password
	wrongLogin := map[string]string{
		"email":    "test@oryn.app",
		"password": "WrongPassword!",
	}
	body, _ = json.Marshal(wrongLogin)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_MeEndpoint(t *testing.T) {
	router, _, jwtMgr := setupAuthTestRouter()

	// 1. Register first
	regPayload := map[string]string{
		"email":       "me@oryn.app",
		"password":    "Password123!",
		"displayName": "Me User",
	}
	body, _ := json.Marshal(regPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// 2. Call /me without auth header or cookie -> 401
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// 3. Call /me with valid Bearer token
	tokenPair, err := jwtMgr.GenerateTokenPair("user-123", "me@oryn.app")
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	// User with ID "user-123" not in memory repo so will return 404
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
