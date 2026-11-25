package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/adapter"
	"github.com/Amirh2006/Oryn/backend/internal/adapter/tmdb"
	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/platform/logger"
	"github.com/Amirh2006/Oryn/backend/internal/platform/resilience"
	"github.com/Amirh2006/Oryn/backend/internal/platform/server"
	"github.com/Amirh2006/Oryn/backend/internal/repository/cache"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/handlers"
	"github.com/Amirh2006/Oryn/backend/internal/usecase"
)

func main() {
	// 1. Load strongly-typed configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 2. Initialize structured logger
	logger.Setup(cfg.App.LogLevel, os.Stdout)

	slog.Info("Starting Oryn Discovery API",
		slog.String("env", cfg.App.Env),
		slog.Int("port", cfg.App.Port),
	)

	// 3. Initialize Redis cache and readiness probe
	redisCache := cache.NewRedisCache(cfg.Redis)
	defer redisCache.Close()

	readinessCheckers := make(map[string]handlers.ReadinessChecker)
	readinessCheckers["redis"] = redisCache

	healthHandler := handlers.NewHealthHandler(readinessCheckers)

	// 4. Initialize MediaProvider (TMDB with Circuit Breaker and Redis Caching)
	cb := resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
		MaxFailures:       5,
		ResetTimeout:      30 * time.Second,
		HalfOpenSuccesses: 2,
	})

	tmdbClient, err := tmdb.NewClient(cfg.TMDB, cb)
	if err != nil {
		slog.Error("Failed to initialize TMDB client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var mediaProvider domain.MediaProvider = adapter.NewCachedMediaProvider(tmdbClient, redisCache)
	mediaUsecase := usecase.NewMediaUsecase(mediaProvider)
	mediaHandler := handlers.NewMediaHandler(mediaUsecase)

	// 5. Initialize HTTP router
	router := http.SetupRouter(&http.RouterConfig{
		Config:        cfg,
		HealthHandler: healthHandler,
		MediaHandler:  mediaHandler,
	})

	// 6. Initialize and run server with graceful shutdown
	srv := server.New(cfg, router)
	if err := srv.Run(); err != nil {
		slog.Error("Server terminated unexpectedly", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
