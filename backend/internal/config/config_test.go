package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Defaults(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_ENV", "development")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "development", cfg.App.Env)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.Equal(t, "info", cfg.App.LogLevel)
	assert.Contains(t, cfg.App.CORSAllowedOrigins, "http://localhost:5173")
}

func TestConfig_ProductionValidation(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_ENV", "production")
	_ = os.Setenv("JWT_SECRET", "short")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters long")
}

func TestConfig_CustomValues(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("APP_ENV", "development")
	_ = os.Setenv("SERVER_PORT", "9090")
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "https://oryn.app, https://admin.oryn.app")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.App.Port)
	assert.Equal(t, "debug", cfg.App.LogLevel)
	assert.Equal(t, []string{"https://oryn.app", "https://admin.oryn.app"}, cfg.App.CORSAllowedOrigins)
}
