package usecase

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/platform/security"
)

type AuthUsecase struct {
	userRepo   domain.UserRepository
	tokenRepo  domain.TokenRepository
	prefRepo   domain.PreferencesRepository
	jwtManager *security.JWTManager
}

func NewAuthUsecase(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	prefRepo domain.PreferencesRepository,
	jwtManager *security.JWTManager,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		prefRepo:   prefRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account and issues tokens
func (u *AuthUsecase) Register(ctx context.Context, email, password, displayName string) (*domain.User, *security.TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, nil, domain.NewAppError(domain.ErrCodeValidation, "invalid email address format", nil)
	}

	if len(password) < 8 {
		return nil, nil, domain.NewAppError(domain.ErrCodeValidation, "password must be at least 8 characters long", nil)
	}

	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}

	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return nil, nil, domain.WrapError(domain.ErrCodeInternal, "failed to hash password", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Initialize default preferences
	_ = u.prefRepo.Upsert(ctx, &domain.UserPreferences{
		UserID:             user.ID,
		Language:           "en",
		Theme:              "dark",
		FavoriteGenres:     []int{},
		PreferredRatingMin: 0.0,
		MaxRuntime:         0,
	})

	tokens, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, domain.WrapError(domain.ErrCodeInternal, "failed to generate tokens", err)
	}

	// Persist refresh token hash
	tokenHash := security.HashToken(tokens.RefreshToken)
	_ = u.tokenRepo.Save(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(tokens.RefreshExpiry),
		Revoked:   false,
	})

	return user, tokens, nil
}

// Login authenticates credentials and returns a new session
func (u *AuthUsecase) Login(ctx context.Context, email, password string) (*domain.User, *security.TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if domain.IsAppError(err, domain.ErrCodeNotFound) {
			return nil, nil, domain.NewAppError(domain.ErrCodeUnauthorized, "invalid email or password", nil)
		}
		return nil, nil, err
	}

	match, err := security.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil || !match {
		return nil, nil, domain.NewAppError(domain.ErrCodeUnauthorized, "invalid email or password", nil)
	}

	tokens, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, domain.WrapError(domain.ErrCodeInternal, "failed to generate tokens", err)
	}

	tokenHash := security.HashToken(tokens.RefreshToken)
	_ = u.tokenRepo.Save(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(tokens.RefreshExpiry),
		Revoked:   false,
	})

	return user, tokens, nil
}

// Refresh handles token rotation
func (u *AuthUsecase) Refresh(ctx context.Context, rawRefreshToken string) (*security.TokenPair, error) {
	claims, err := u.jwtManager.ValidateToken(rawRefreshToken)
	if err != nil {
		return nil, domain.NewAppError(domain.ErrCodeUnauthorized, "invalid or expired refresh token", nil)
	}

	tokenHash := security.HashToken(rawRefreshToken)
	storedToken, err := u.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil || storedToken.Revoked || time.Now().UTC().After(storedToken.ExpiresAt) {
		return nil, domain.NewAppError(domain.ErrCodeUnauthorized, "refresh token is revoked or expired", nil)
	}

	// Revoke old token
	_ = u.tokenRepo.RevokeByHash(ctx, tokenHash)

	// Issue rotated pair
	newTokens, err := u.jwtManager.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return nil, domain.WrapError(domain.ErrCodeInternal, "failed to issue new tokens", err)
	}

	newTokenHash := security.HashToken(newTokens.RefreshToken)
	_ = u.tokenRepo.Save(ctx, &domain.RefreshToken{
		UserID:    claims.UserID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().UTC().Add(newTokens.RefreshExpiry),
		Revoked:   false,
	})

	return newTokens, nil
}

// Logout revokes session refresh token
func (u *AuthUsecase) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}
	tokenHash := security.HashToken(rawRefreshToken)
	return u.tokenRepo.RevokeByHash(ctx, tokenHash)
}

// GetMe fetches authenticated user profile
func (u *AuthUsecase) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, userID)
}
