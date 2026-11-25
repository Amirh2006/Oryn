package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all strongly-typed configuration for the application.
type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	Auth       AuthConfig
	TMDB       TMDBConfig
	Cloudflare CloudflareConfig
}

type AppConfig struct {
	Env                string
	Name               string
	Port               int
	Host               string
	LogLevel           string
	CORSAllowedOrigins []string
	ShutdownTimeout    time.Duration
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AuthConfig struct {
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	CookieDomain     string
	CookieSecure     bool
}

type TMDBConfig struct {
	APIKey           string
	BaseURL          string
	ImageBaseURL     string
	OutboundProxyURL string
}

type CloudflareConfig struct {
	AccountID         string
	APIToken          string
	MainModel         string
	LightModel        string
	DailyNeuronBudget int
}

// Load reads configuration from environment variables, optionally loading from a .env file.
func Load() (*Config, error) {
	// Attempt loading .env file if it exists, ignore if absent
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:                getEnv("APP_ENV", "development"),
			Name:               getEnv("APP_NAME", "oryn"),
			Port:               getEnvAsInt("SERVER_PORT", 8080),
			Host:               getEnv("SERVER_HOST", "0.0.0.0"),
			LogLevel:           getEnv("LOG_LEVEL", "info"),
			CORSAllowedOrigins: splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
			ShutdownTimeout:    time.Duration(getEnvAsInt("SHUTDOWN_TIMEOUT_SECONDS", 10)) * time.Second,
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", ""),
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "oryn"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", "oryn_db"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret:        getEnv("JWT_SECRET", ""),
			JWTAccessExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			JWTRefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", 168*time.Hour),
			CookieDomain:     getEnv("COOKIE_DOMAIN", "localhost"),
			CookieSecure:     getEnvAsBool("COOKIE_SECURE", false),
		},
		TMDB: TMDBConfig{
			APIKey:           getEnv("TMDB_API_KEY", ""),
			BaseURL:          getEnv("TMDB_BASE_URL", "https://api.themoviedb.org/3"),
			ImageBaseURL:     getEnv("TMDB_IMAGE_BASE_URL", "https://image.tmdb.org/t/p"),
			OutboundProxyURL: getEnv("OUTBOUND_PROXY_URL", ""),
		},
		Cloudflare: CloudflareConfig{
			AccountID:         getEnv("CLOUDFLARE_ACCOUNT_ID", ""),
			APIToken:          getEnv("CLOUDFLARE_API_TOKEN", ""),
			MainModel:         getEnv("AI_MAIN_MODEL", "@cf/meta/llama-3.1-8b-instruct"),
			LightModel:        getEnv("AI_LIGHT_MODEL", "@cf/meta/llama-3.1-8b-instruct-fast"),
			DailyNeuronBudget: getEnvAsInt("AI_DAILY_NEURON_BUDGET", 10000),
		},
	}

	// In test environment, set a valid default JWT secret if missing
	if cfg.App.Env == "test" && cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = "test_environment_secret_minimum_32_characters_key"
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks for mandatory configuration values.
func (c *Config) Validate() error {
	var errs []string

	if c.App.Port <= 0 || c.App.Port > 65535 {
		errs = append(errs, "SERVER_PORT must be between 1 and 65535")
	}

	// In production, ensure JWTSecret meets high entropy requirements
	if c.App.Env == "production" {
		if len(c.Auth.JWTSecret) < 32 {
			errs = append(errs, "JWT_SECRET must be at least 32 characters long in production")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Helper utility functions
func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
