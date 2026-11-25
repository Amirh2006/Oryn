package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	query := `
		INSERT INTO users (id, email, password_hash, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query, u.ID, u.Email, u.PasswordHash, u.DisplayName, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.NewAppError(domain.ErrCodeConflict, "user with this email already exists", nil)
		}
		return domain.WrapError(domain.ErrCodeInternal, "failed to insert user", err)
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	u := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapError(domain.ErrCodeInternal, "failed to query user by id", err)
	}

	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	u := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.WrapError(domain.ErrCodeInternal, "failed to query user by email", err)
	}

	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	u.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE users
		SET display_name = $1, updated_at = $2
		WHERE id = $3
	`

	res, err := r.db.ExecContext(ctx, query, u.DisplayName, u.UpdatedAt, u.ID)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to update user", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}
