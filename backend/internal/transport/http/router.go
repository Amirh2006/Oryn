package http

import (
	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/Amirh2006/Oryn/backend/internal/platform/security"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/handlers"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RouterConfig contains dependencies for setting up HTTP routes
type RouterConfig struct {
	Config         *config.Config
	HealthHandler  *handlers.HealthHandler
	MediaHandler   *handlers.MediaHandler
	AuthHandler    *handlers.AuthHandler
	LibraryHandler *handlers.LibraryHandler
	JWTManager     *security.JWTManager
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

		// Media Discovery Endpoints
		if cfg.MediaHandler != nil {
			media := apiV1.Group("/media")
			{
				media.GET("/trending", cfg.MediaHandler.GetTrending)
				media.GET("/popular", cfg.MediaHandler.GetPopular)
				media.GET("/top-rated", cfg.MediaHandler.GetTopRated)
				media.GET("/search", cfg.MediaHandler.Search)
				media.GET("/discover", cfg.MediaHandler.Discover)
				media.GET("/genres", cfg.MediaHandler.GetGenres)
				media.GET("/:type/:id", cfg.MediaHandler.GetDetails)
				media.GET("/:type/:id/similar", cfg.MediaHandler.GetSimilar)
			}
		}

		// Authentication Endpoints
		if cfg.AuthHandler != nil {
			auth := apiV1.Group("/auth")
			{
				auth.POST("/register", cfg.AuthHandler.Register)
				auth.POST("/login", cfg.AuthHandler.Login)
				auth.POST("/refresh", cfg.AuthHandler.Refresh)
				auth.POST("/logout", cfg.AuthHandler.Logout)

				if cfg.JWTManager != nil {
					auth.GET("/me", middleware.Auth(cfg.JWTManager), cfg.AuthHandler.Me)
				}
			}
		}

		// Authenticated User Preferences & Personal Library
		if cfg.LibraryHandler != nil && cfg.JWTManager != nil {
			authRequired := middleware.Auth(cfg.JWTManager)

			users := apiV1.Group("/users")
			users.Use(authRequired)
			{
				users.GET("/preferences", cfg.LibraryHandler.GetPreferences)
				users.PUT("/preferences", cfg.LibraryHandler.UpdatePreferences)
			}

			library := apiV1.Group("/library")
			library.Use(authRequired)
			{
				library.GET("/favorites", cfg.LibraryHandler.GetFavorites)
				library.POST("/favorites", cfg.LibraryHandler.AddFavorite)
				library.DELETE("/favorites/:type/:id", cfg.LibraryHandler.RemoveFavorite)
				library.GET("/favorites/:type/:id/check", cfg.LibraryHandler.CheckFavorite)

				library.GET("/collections", cfg.LibraryHandler.GetCollections)
				library.POST("/collections", cfg.LibraryHandler.CreateCollection)
				library.GET("/collections/:id", cfg.LibraryHandler.GetCollectionDetails)
				library.DELETE("/collections/:id", cfg.LibraryHandler.DeleteCollection)
				library.POST("/collections/:id/items", cfg.LibraryHandler.AddCollectionItem)
				library.DELETE("/collections/:id/items/:itemId", cfg.LibraryHandler.RemoveCollectionItem)
				library.PATCH("/collections/:id/items/:itemId/status", cfg.LibraryHandler.UpdateWatchStatus)
			}
		}
	}

	return r
}
