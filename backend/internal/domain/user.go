package domain

import (
	"context"
	"time"
)

// User represents an authenticated account
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"displayName"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// RefreshToken represents a persisted session refresh token
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserPreferences represents user profile customizations
type UserPreferences struct {
	UserID             string    `json:"userId"`
	Language           string    `json:"language"`
	Theme              string    `json:"theme"`
	FavoriteGenres     []int     `json:"favoriteGenres"`
	PreferredRatingMin float64   `json:"preferredRatingMin"`
	MaxRuntime         int       `json:"maxRuntime"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// UserRepository defines the persistence interface for users
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}

// TokenRepository defines persistence operations for session refresh tokens
type TokenRepository interface {
	Save(ctx context.Context, token *RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeByHash(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

// PreferencesRepository defines persistence operations for user preferences
type PreferencesRepository interface {
	GetByUserID(ctx context.Context, userID string) (*UserPreferences, error)
	Upsert(ctx context.Context, prefs *UserPreferences) error
}
