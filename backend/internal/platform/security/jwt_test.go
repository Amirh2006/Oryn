package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	secret := "test-secret-at-least-32-bytes-long-for-hmac-sha256"
	manager := NewJWTManager(secret, 15*time.Minute, 24*time.Hour)

	userID := "usr-12345-uuid"
	email := "user@oryn.app"

	tokens, err := manager.GenerateTokenPair(userID, email)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	// Validate Access Token
	claims, err := manager.ValidateToken(tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "oryn-api", claims.Issuer)

	// Validate Refresh Token
	refClaims, err := manager.ValidateToken(tokens.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, refClaims.UserID)
	assert.Equal(t, "oryn-api-refresh", refClaims.Issuer)
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	secret := "test-secret-at-least-32-bytes-long-for-hmac-sha256"
	manager := NewJWTManager(secret, -1*time.Minute, -1*time.Minute)

	tokens, err := manager.GenerateTokenPair("usr-expired", "expired@oryn.app")
	require.NoError(t, err)

	_, err = manager.ValidateToken(tokens.AccessToken)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestJWTManager_InvalidSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-key-one-32-characters-length!!", 15*time.Minute, 24*time.Hour)
	manager2 := NewJWTManager("secret-key-two-32-characters-length!!", 15*time.Minute, 24*time.Hour)

	tokens, err := manager1.GenerateTokenPair("usr-test", "test@oryn.app")
	require.NoError(t, err)

	_, err = manager2.ValidateToken(tokens.AccessToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestHashToken(t *testing.T) {
	token := "some-sample-token-string"
	hash1 := HashToken(token)
	hash2 := HashToken(token)

	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 64) // SHA-256 hex string is 64 characters
}
