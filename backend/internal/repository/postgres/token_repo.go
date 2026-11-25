package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/google/uuid"
)

type TokenRepo struct {
	db *DB
}

func NewTokenRepo(db *DB) *TokenRepo {
	return &TokenRepo{db: db}
}

func (r *TokenRepo) Save(ctx context.Context, t *domain.RefreshToken) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now().UTC()

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query, t.ID, t.UserID, t.TokenHash, t.ExpiresAt, t.Revoked, t.CreatedAt)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to insert refresh token", err)
	}

	return nil
}

func (r *TokenRepo) GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	t := &domain.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.Revoked, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapError(domain.ErrCodeInternal, "failed to query token", err)
	}

	return t, nil
}

func (r *TokenRepo) RevokeByHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE token_hash = $1
	`

	_, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to revoke token", err)
	}

	return nil
}

func (r *TokenRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to revoke user tokens", err)
	}

	return nil
}
