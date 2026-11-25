package http

import (
	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/handlers"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RouterConfig contains dependencies for setting up HTTP routes
type RouterConfig struct {
	Config        *config.Config
	HealthHandler *handlers.HealthHandler
}

// SetupRouter initializes the Gin engine with all global middlewares and routes
func SetupRouter(cfg *RouterConfig) *gin.Engine {
	if cfg.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// Global middlewares
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.CORS(cfg.Config.App.CORSAllowedOrigins))
	r.Use(middleware.Gzip())

	// Kubernetes and container lifecycle probes
	r.GET("/healthz", cfg.HealthHandler.Healthz)
	r.GET("/readyz", cfg.HealthHandler.Readyz)

	// API versioned group
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}

	return r
}
