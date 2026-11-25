package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/transport/http/response"
	"github.com/gin-gonic/gin"
)

// ReadinessChecker defines the interface for checking readiness of system dependencies (DB, Redis, etc.)
type ReadinessChecker interface {
	CheckReadiness(ctx context.Context) error
}

// HealthHandler manages liveness and readiness probes
type HealthHandler struct {
	startTime time.Time
	checkers  map[string]ReadinessChecker
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(checkers map[string]ReadinessChecker) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		checkers:  checkers,
	}
}

// Healthz responds to kubernetes/container liveness probe
func (h *HealthHandler) Healthz(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{
		"status":  "healthy",
		"uptime":  time.Since(h.startTime).String(),
		"service": "oryn-api",
	})
}

// Readyz responds to kubernetes/container readiness probe
func (h *HealthHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	failures := make(map[string]string)
	for name, checker := range h.checkers {
		if checker != nil {
			if err := checker.CheckReadiness(ctx); err != nil {
				failures[name] = err.Error()
			}
		}
	}

	if len(failures) > 0 {
		response.Error(c, http.StatusServiceUnavailable, "DEPENDENCY_NOT_READY", "one or more dependencies are not ready", failures)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"status": "ready",
		"checks": "all passed",
	})
}
