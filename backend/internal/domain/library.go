package domain

import (
	"context"
	"time"
)

type WatchStatus string

const (
	WatchStatusPlanToWatch WatchStatus = "plan_to_watch"
	WatchStatusWatching    WatchStatus = "watching"
	WatchStatusWatched     WatchStatus = "watched"
)

// FavoriteItem represents a media item bookmarked by the user
type FavoriteItem struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	ProviderName    string    `json:"providerName"`
	ProviderMediaID int       `json:"providerMediaId"`
	MediaType       string    `json:"mediaType"`
	Title           string    `json:"title"`
	PosterURL       string    `json:"posterUrl"`
	Rating          float64   `json:"rating"`
	ReleaseYear     int       `json:"releaseYear"`
	CreatedAt       time.Time `json:"createdAt"`
}

// CustomList represents a user-created collection
type CustomList struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	IsPublic    bool      `json:"isPublic"`
	ItemsCount  int       `json:"itemsCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CustomListItem represents an entry in a custom collection
type CustomListItem struct {
	ID              string      `json:"id"`
	ListID          string      `json:"listId"`
	ProviderName    string      `json:"providerName"`
	ProviderMediaID int         `json:"providerMediaId"`
	MediaType       string      `json:"mediaType"`
	Title           string      `json:"title"`
	PosterURL       string      `json:"posterUrl"`
	WatchStatus     WatchStatus `json:"watchStatus"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

// LibraryRepository defines database operations for personal library
type LibraryRepository interface {
	AddFavorite(ctx context.Context, fav *FavoriteItem) error
	RemoveFavorite(ctx context.Context, userID, mediaType string, providerMediaID int) error
	GetFavorites(ctx context.Context, userID string, page, limit int) ([]FavoriteItem, int64, error)
	IsFavorite(ctx context.Context, userID, mediaType string, providerMediaID int) (bool, error)
	CreateCustomList(ctx context.Context, list *CustomList) error
	GetCustomLists(ctx context.Context, userID string) ([]CustomList, error)
	GetCustomListByID(ctx context.Context, listID, userID string) (*CustomList, []CustomListItem, error)
	DeleteCustomList(ctx context.Context, listID, userID string) error
	AddCustomListItem(ctx context.Context, userID string, item *CustomListItem) error
	RemoveCustomListItem(ctx context.Context, userID, itemID string) error
	UpdateWatchStatus(ctx context.Context, userID, itemID string, status WatchStatus) error
}
