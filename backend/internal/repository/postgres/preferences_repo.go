package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
)

type PreferencesRepo struct {
	db *DB
}

func NewPreferencesRepo(db *DB) *PreferencesRepo {
	return &PreferencesRepo{db: db}
}

func (r *PreferencesRepo) GetByUserID(ctx context.Context, userID string) (*domain.UserPreferences, error) {
	query := `
		SELECT user_id, language, theme, favorite_genres, preferred_rating_min, max_runtime, created_at, updated_at
		FROM user_preferences
		WHERE user_id = $1
	`

	p := &domain.UserPreferences{}
	var genresJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&p.UserID, &p.Language, &p.Theme, &genresJSON, &p.PreferredRatingMin, &p.MaxRuntime, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default preferences if not configured yet
			return &domain.UserPreferences{
				UserID:             userID,
				Language:           "en",
				Theme:              "dark",
				FavoriteGenres:     []int{},
				PreferredRatingMin: 0.0,
				MaxRuntime:         0,
			}, nil
		}
		return nil, domain.WrapError(domain.ErrCodeInternal, "failed to query user preferences", err)
	}

	if len(genresJSON) > 0 {
		_ = json.Unmarshal(genresJSON, &p.FavoriteGenres)
	}

	return p, nil
}

func (r *PreferencesRepo) Upsert(ctx context.Context, p *domain.UserPreferences) error {
	now := time.Now().UTC()
	p.UpdatedAt = now
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}

	genresJSON, err := json.Marshal(p.FavoriteGenres)
	if err != nil {
		genresJSON = []byte("[]")
	}

	query := `
		INSERT INTO user_preferences (user_id, language, theme, favorite_genres, preferred_rating_min, max_runtime, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE
		SET language = EXCLUDED.language,
		    theme = EXCLUDED.theme,
		    favorite_genres = EXCLUDED.favorite_genres,
		    preferred_rating_min = EXCLUDED.preferred_rating_min,
		    max_runtime = EXCLUDED.max_runtime,
		    updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.ExecContext(ctx, query,
		p.UserID, p.Language, p.Theme, genresJSON, p.PreferredRatingMin, p.MaxRuntime, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to upsert user preferences", err)
	}

	return nil
}
