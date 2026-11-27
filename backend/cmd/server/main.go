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
	"github.com/Amirh2006/Oryn/backend/internal/platform/security"
	"github.com/Amirh2006/Oryn/backend/internal/platform/server"
	"github.com/Amirh2006/Oryn/backend/internal/repository/cache"
	"github.com/Amirh2006/Oryn/backend/internal/repository/postgres"
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

	slog.Info("Starting Oryn API Server",
		slog.String("env", cfg.App.Env),
		slog.Int("port", cfg.App.Port),
	)

	// 3. Initialize Redis cache and readiness probe
	redisCache := cache.NewRedisCache(cfg.Redis)
	defer redisCache.Close()

	readinessCheckers := make(map[string]handlers.ReadinessChecker)
	readinessCheckers["redis"] = redisCache

	// 4. Initialize PostgreSQL connection pool and readiness probe
	postgresDB, err := postgres.NewDB(cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize PostgreSQL connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer postgresDB.Close()
	readinessCheckers["postgres"] = postgresDB

	healthHandler := handlers.NewHealthHandler(readinessCheckers)

	// 5. Initialize Security & Repositories
	jwtManager := security.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTAccessExpiry, cfg.Auth.JWTRefreshExpiry)

	userRepo := postgres.NewUserRepo(postgresDB)
	tokenRepo := postgres.NewTokenRepo(postgresDB)
	prefRepo := postgres.NewPreferencesRepo(postgresDB)
	libraryRepo := postgres.NewLibraryRepo(postgresDB)

	// 6. Initialize Use Cases & Handlers
	authUsecase := usecase.NewAuthUsecase(userRepo, tokenRepo, prefRepo, jwtManager)
	authHandler := handlers.NewAuthHandler(authUsecase, cfg.Auth)

	libraryUsecase := usecase.NewLibraryUsecase(libraryRepo, prefRepo)
	libraryHandler := handlers.NewLibraryHandler(libraryUsecase)

	// 7. Initialize MediaProvider (TMDB with Circuit Breaker and Redis Caching)
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

	// 8. Initialize HTTP router
	router := http.SetupRouter(&http.RouterConfig{
		Config:         cfg,
		HealthHandler:  healthHandler,
		MediaHandler:   mediaHandler,
		AuthHandler:    authHandler,
		LibraryHandler: libraryHandler,
		JWTManager:     jwtManager,
	})

	// 9. Initialize and run server with graceful shutdown
	srv := server.New(cfg, router)
	if err := srv.Run(); err != nil {
		slog.Error("Server terminated unexpectedly", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
