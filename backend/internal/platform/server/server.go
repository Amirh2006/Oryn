package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/config"
)

// Server wraps http.Server with graceful shutdown capabilities
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

// New creates a new HTTP Server instance
func New(cfg *config.Config, handler http.Handler) *Server {
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		cfg: cfg,
	}
}

// Run starts the HTTP server and blocks until an interrupt signal is received
func (s *Server) Run() error {
	shutdownErr := make(chan error, 1)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		slog.Info("Server is shutting down...", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.App.ShutdownTimeout)
		defer cancel()

		shutdownErr <- s.httpServer.Shutdown(ctx)
	}()

	slog.Info("Oryn API Server listening",
		slog.String("address", s.httpServer.Addr),
		slog.String("environment", s.cfg.App.Env),
	)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server listen error: %w", err)
	}

	err := <-shutdownErr
	if err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	slog.Info("Server stopped cleanly")
	return nil
}
