package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgon2id_HashAndVerify(t *testing.T) {
	password := "OrynSuperSecurePassword2026!"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")

	// Correct password
	match, err := ComparePasswordAndHash(password, hash)
	require.NoError(t, err)
	assert.True(t, match)

	// Incorrect password
	match, err = ComparePasswordAndHash("WrongPassword123", hash)
	require.NoError(t, err)
	assert.False(t, match)
}

func TestArgon2id_InvalidHash(t *testing.T) {
	match, err := ComparePasswordAndHash("pass", "invalid-hash-string")
	assert.Error(t, err)
	assert.False(t, match)
}
