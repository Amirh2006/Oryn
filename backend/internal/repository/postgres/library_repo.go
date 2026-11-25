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

type LibraryRepo struct {
	db *DB
}

func NewLibraryRepo(db *DB) *LibraryRepo {
	return &LibraryRepo{db: db}
}

// AddFavorite adds an item to user's favorites
func (r *LibraryRepo) AddFavorite(ctx context.Context, fav *domain.FavoriteItem) error {
	if fav.ID == "" {
		fav.ID = uuid.New().String()
	}
	fav.CreatedAt = time.Now().UTC()

	query := `
		INSERT INTO favorites (id, user_id, provider_name, provider_media_id, media_type, title, poster_url, rating, release_year, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, provider_name, provider_media_id, media_type) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query,
		fav.ID, fav.UserID, fav.ProviderName, fav.ProviderMediaID, fav.MediaType,
		fav.Title, fav.PosterURL, fav.Rating, fav.ReleaseYear, fav.CreatedAt,
	)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to add favorite", err)
	}

	return nil
}

// RemoveFavorite removes a media item from user's favorites
func (r *LibraryRepo) RemoveFavorite(ctx context.Context, userID, mediaType string, providerMediaID int) error {
	query := `
		DELETE FROM favorites
		WHERE user_id = $1 AND media_type = $2 AND provider_media_id = $3
	`

	res, err := r.db.ExecContext(ctx, query, userID, mediaType, providerMediaID)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to remove favorite", err)
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

// GetFavorites returns paginated list of user's favorites
func (r *LibraryRepo) GetFavorites(ctx context.Context, userID string, page, limit int) ([]domain.FavoriteItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, domain.WrapError(domain.ErrCodeInternal, "failed to count favorites", err)
	}

	query := `
		SELECT id, user_id, provider_name, provider_media_id, media_type, title, poster_url, rating, release_year, created_at
		FROM favorites
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, domain.WrapError(domain.ErrCodeInternal, "failed to query favorites", err)
	}
	defer rows.Close()

	var items []domain.FavoriteItem
	for rows.Next() {
		var itm domain.FavoriteItem
		if err := rows.Scan(
			&itm.ID, &itm.UserID, &itm.ProviderName, &itm.ProviderMediaID, &itm.MediaType,
			&itm.Title, &itm.PosterURL, &itm.Rating, &itm.ReleaseYear, &itm.CreatedAt,
		); err != nil {
			return nil, 0, domain.WrapError(domain.ErrCodeInternal, "failed to scan favorite item", err)
		}
		items = append(items, itm)
	}

	return items, total, nil
}

// IsFavorite checks whether a media item is bookmarked by user
func (r *LibraryRepo) IsFavorite(ctx context.Context, userID, mediaType string, providerMediaID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM favorites
			WHERE user_id = $1 AND media_type = $2 AND provider_media_id = $3
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, mediaType, providerMediaID).Scan(&exists)
	if err != nil {
		return false, domain.WrapError(domain.ErrCodeInternal, "failed to check favorite status", err)
	}
	return exists, nil
}

// CreateCustomList creates a new collection
func (r *LibraryRepo) CreateCustomList(ctx context.Context, list *domain.CustomList) error {
	if list.ID == "" {
		list.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	list.CreatedAt = now
	list.UpdatedAt = now

	query := `
		INSERT INTO custom_lists (id, user_id, title, description, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		list.ID, list.UserID, list.Title, list.Description, list.IsPublic, list.CreatedAt, list.UpdatedAt,
	)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to create custom list", err)
	}

	return nil
}

// GetCustomLists returns all collections owned by user
func (r *LibraryRepo) GetCustomLists(ctx context.Context, userID string) ([]domain.CustomList, error) {
	query := `
		SELECT cl.id, cl.user_id, cl.title, cl.description, cl.is_public, cl.created_at, cl.updated_at,
		       COUNT(cli.id) AS items_count
		FROM custom_lists cl
		LEFT JOIN custom_list_items cli ON cl.id = cli.list_id
		WHERE cl.user_id = $1
		GROUP BY cl.id
		ORDER BY cl.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, domain.WrapError(domain.ErrCodeInternal, "failed to query custom lists", err)
	}
	defer rows.Close()

	var lists []domain.CustomList
	for rows.Next() {
		var l domain.CustomList
		var desc sql.NullString
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.Title, &desc, &l.IsPublic, &l.CreatedAt, &l.UpdatedAt, &l.ItemsCount,
		); err != nil {
			return nil, domain.WrapError(domain.ErrCodeInternal, "failed to scan custom list", err)
		}
		if desc.Valid {
			l.Description = desc.String
		}
		lists = append(lists, l)
	}

	return lists, nil
}

// GetCustomListByID returns collection details and its items
func (r *LibraryRepo) GetCustomListByID(ctx context.Context, listID, userID string) (*domain.CustomList, []domain.CustomListItem, error) {
	listQuery := `
		SELECT id, user_id, title, description, is_public, created_at, updated_at
		FROM custom_lists
		WHERE id = $1 AND (user_id = $2 OR is_public = TRUE)
	`

	l := &domain.CustomList{}
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, listQuery, listID, userID).Scan(
		&l.ID, &l.UserID, &l.Title, &desc, &l.IsPublic, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, domain.WrapError(domain.ErrCodeInternal, "failed to query list by id", err)
	}
	if desc.Valid {
		l.Description = desc.String
	}

	itemsQuery := `
		SELECT id, list_id, provider_name, provider_media_id, media_type, title, poster_url, watch_status, created_at, updated_at
		FROM custom_list_items
		WHERE list_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, listID)
	if err != nil {
		return nil, nil, domain.WrapError(domain.ErrCodeInternal, "failed to query list items", err)
	}
	defer rows.Close()

	var items []domain.CustomListItem
	for rows.Next() {
		var itm domain.CustomListItem
		if err := rows.Scan(
			&itm.ID, &itm.ListID, &itm.ProviderName, &itm.ProviderMediaID, &itm.MediaType,
			&itm.Title, &itm.PosterURL, &itm.WatchStatus, &itm.CreatedAt, &itm.UpdatedAt,
		); err != nil {
			return nil, nil, domain.WrapError(domain.ErrCodeInternal, "failed to scan list item", err)
		}
		items = append(items, itm)
	}
	l.ItemsCount = len(items)

	return l, items, nil
}

// DeleteCustomList deletes collection ensuring user ownership
func (r *LibraryRepo) DeleteCustomList(ctx context.Context, listID, userID string) error {
	query := `
		DELETE FROM custom_lists
		WHERE id = $1 AND user_id = $2
	`

	res, err := r.db.ExecContext(ctx, query, listID, userID)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to delete custom list", err)
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

// AddCustomListItem adds an item to custom list verifying list belongs to user
func (r *LibraryRepo) AddCustomListItem(ctx context.Context, userID string, item *domain.CustomListItem) error {
	// Verify list ownership
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM custom_lists WHERE id = $1 AND user_id = $2)`
	if err := r.db.QueryRowContext(ctx, checkQuery, item.ListID, userID).Scan(&exists); err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to verify list ownership", err)
	}
	if !exists {
		return domain.ErrForbidden
	}

	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	if item.WatchStatus == "" {
		item.WatchStatus = domain.WatchStatusPlanToWatch
	}

	query := `
		INSERT INTO custom_list_items (id, list_id, provider_name, provider_media_id, media_type, title, poster_url, watch_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (list_id, provider_name, provider_media_id, media_type) DO UPDATE
		SET watch_status = EXCLUDED.watch_status,
		    updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.ListID, item.ProviderName, item.ProviderMediaID, item.MediaType,
		item.Title, item.PosterURL, item.WatchStatus, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23514" {
			return domain.NewAppError(domain.ErrCodeValidation, "invalid watch_status (must be watched, watching, or plan_to_watch)", nil)
		}
		return domain.WrapError(domain.ErrCodeInternal, "failed to insert list item", err)
	}

	return nil
}

// RemoveCustomListItem deletes an item from custom list
func (r *LibraryRepo) RemoveCustomListItem(ctx context.Context, userID, itemID string) error {
	query := `
		DELETE FROM custom_list_items cli
		USING custom_lists cl
		WHERE cli.list_id = cl.id AND cli.id = $1 AND cl.user_id = $2
	`

	res, err := r.db.ExecContext(ctx, query, itemID, userID)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to delete list item", err)
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

// UpdateWatchStatus updates watch status of an item ensuring user ownership
func (r *LibraryRepo) UpdateWatchStatus(ctx context.Context, userID, itemID string, status domain.WatchStatus) error {
	if status != domain.WatchStatusWatched && status != domain.WatchStatusWatching && status != domain.WatchStatusPlanToWatch {
		return domain.NewAppError(domain.ErrCodeValidation, "invalid watch_status value", nil)
	}

	query := `
		UPDATE custom_list_items cli
		SET watch_status = $1, updated_at = $2
		FROM custom_lists cl
		WHERE cli.list_id = cl.id AND cli.id = $3 AND cl.user_id = $4
	`

	res, err := r.db.ExecContext(ctx, query, status, time.Now().UTC(), itemID, userID)
	if err != nil {
		return domain.WrapError(domain.ErrCodeInternal, "failed to update watch status", err)
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
