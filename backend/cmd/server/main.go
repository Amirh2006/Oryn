package main

import (
	"log/slog"
	"os"

	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/Amirh2006/Oryn/backend/internal/platform/logger"
	"github.com/Amirh2006/Oryn/backend/internal/platform/server"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/handlers"
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

	// 3. Initialize dependencies and readiness checkers
	// In Phase 1, we start with empty map; Phase 2 & 3 register Redis and Postgres
	readinessCheckers := make(map[string]handlers.ReadinessChecker)
	healthHandler := handlers.NewHealthHandler(readinessCheckers)

	// 4. Initialize HTTP router
	router := http.SetupRouter(&http.RouterConfig{
		Config:        cfg,
		HealthHandler: healthHandler,
	})

	// 5. Initialize and run server with graceful shutdown
	srv := server.New(cfg, router)
	if err := srv.Run(); err != nil {
		slog.Error("Server terminated unexpectedly", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
