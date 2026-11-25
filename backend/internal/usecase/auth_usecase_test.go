package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/platform/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	m.users[u.Email] = u
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	m.users[u.Email] = u
	return nil
}

type mockTokenRepo struct {
	tokens map[string]*domain.RefreshToken
}

func (m *mockTokenRepo) Save(ctx context.Context, t *domain.RefreshToken) error {
	m.tokens[t.TokenHash] = t
	return nil
}
func (m *mockTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	if t, ok := m.tokens[hash]; ok {
		return t, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockTokenRepo) RevokeByHash(ctx context.Context, hash string) error {
	if t, ok := m.tokens[hash]; ok {
		t.Revoked = true
	}
	return nil
}
func (m *mockTokenRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
		}
	}
	return nil
}

type mockPrefRepo struct {
	prefs map[string]*domain.UserPreferences
}

func (m *mockPrefRepo) GetByUserID(ctx context.Context, userID string) (*domain.UserPreferences, error) {
	if p, ok := m.prefs[userID]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockPrefRepo) Upsert(ctx context.Context, p *domain.UserPreferences) error {
	m.prefs[p.UserID] = p
	return nil
}

func TestAuthUsecase_RegisterAndLogin(t *testing.T) {
	uRepo := &mockUserRepo{users: make(map[string]*domain.User)}
	tRepo := &mockTokenRepo{tokens: make(map[string]*domain.RefreshToken)}
	pRepo := &mockPrefRepo{prefs: make(map[string]*domain.UserPreferences)}
	jwtMgr := security.NewJWTManager("test-super-secret-key-32-chars-length!", 15*time.Minute, 24*time.Hour)

	uc := NewAuthUsecase(uRepo, tRepo, pRepo, jwtMgr)
	ctx := context.Background()

	// 1. Register
	user, tokens, err := uc.Register(ctx, "amir@oryn.app", "StrongPassword123!", "Amir")
	require.NoError(t, err)
	assert.Equal(t, "amir@oryn.app", user.Email)
	assert.NotEmpty(t, tokens.AccessToken)

	// 2. Login with correct password
	loggedUser, loginTokens, err := uc.Login(ctx, "amir@oryn.app", "StrongPassword123!")
	require.NoError(t, err)
	assert.Equal(t, user.ID, loggedUser.ID)
	assert.NotEmpty(t, loginTokens.AccessToken)

	// 3. Login with wrong password
	_, _, err = uc.Login(ctx, "amir@oryn.app", "WrongPassword!")
	assert.Error(t, err)

	// 4. Token Refresh
	newTokens, err := uc.Refresh(ctx, loginTokens.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newTokens.AccessToken)

	// 5. Old token should now be revoked
	_, err = uc.Refresh(ctx, loginTokens.RefreshToken)
	assert.Error(t, err)
}
