package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockChecker struct {
	err error
}

func (m *mockChecker) CheckReadiness(ctx context.Context) error {
	return m.err
}

func TestHealthHandler_Healthz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.Healthz(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestHealthHandler_Readyz_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checkers := map[string]ReadinessChecker{
		"postgres": &mockChecker{err: nil},
		"redis":    &mockChecker{err: nil},
	}
	h := NewHealthHandler(checkers)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Readyz(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ready")
}

func TestHealthHandler_Readyz_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checkers := map[string]ReadinessChecker{
		"postgres": &mockChecker{err: errors.New("connection refused")},
	}
	h := NewHealthHandler(checkers)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Readyz(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "DEPENDENCY_NOT_READY")
}
