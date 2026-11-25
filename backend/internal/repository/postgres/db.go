package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/config"
	_ "github.com/lib/pq"
)

// DB wraps sql.DB and implements handlers.ReadinessChecker
type DB struct {
	*sql.DB
}

// NewDB initializes PostgreSQL connection pool
func NewDB(cfg config.DatabaseConfig) (*DB, error) {
	connStr := cfg.URL
	if connStr == "" {
		connStr = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Production connection pool tuning
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	slog.Info("PostgreSQL connection pool configured",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("dbname", cfg.DBName),
	)

	return &DB{DB: db}, nil
}

// CheckReadiness implements handlers.ReadinessChecker
func (d *DB) CheckReadiness(ctx context.Context) error {
	if d.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return d.PingContext(ctx)
}
