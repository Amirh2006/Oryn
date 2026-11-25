package usecase

import (
	"context"
	"strings"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
)

type LibraryUsecase struct {
	libraryRepo domain.LibraryRepository
	prefRepo    domain.PreferencesRepository
}

func NewLibraryUsecase(libraryRepo domain.LibraryRepository, prefRepo domain.PreferencesRepository) *LibraryUsecase {
	return &LibraryUsecase{
		libraryRepo: libraryRepo,
		prefRepo:    prefRepo,
	}
}

// User Preferences
func (u *LibraryUsecase) GetPreferences(ctx context.Context, userID string) (*domain.UserPreferences, error) {
	return u.prefRepo.GetByUserID(ctx, userID)
}

func (u *LibraryUsecase) UpdatePreferences(ctx context.Context, prefs *domain.UserPreferences) error {
	if prefs.Language == "" {
		prefs.Language = "en"
	}
	if prefs.Theme == "" {
		prefs.Theme = "dark"
	}
	return u.prefRepo.Upsert(ctx, prefs)
}

// Favorites
func (u *LibraryUsecase) AddFavorite(ctx context.Context, fav *domain.FavoriteItem) error {
	if fav.UserID == "" || fav.ProviderMediaID == 0 || fav.MediaType == "" {
		return domain.NewAppError(domain.ErrCodeValidation, "invalid favorite media parameters", nil)
	}
	return u.libraryRepo.AddFavorite(ctx, fav)
}

func (u *LibraryUsecase) RemoveFavorite(ctx context.Context, userID, mediaType string, providerMediaID int) error {
	return u.libraryRepo.RemoveFavorite(ctx, userID, mediaType, providerMediaID)
}

func (u *LibraryUsecase) GetFavorites(ctx context.Context, userID string, page, limit int) ([]domain.FavoriteItem, int64, error) {
	return u.libraryRepo.GetFavorites(ctx, userID, page, limit)
}

func (u *LibraryUsecase) IsFavorite(ctx context.Context, userID, mediaType string, providerMediaID int) (bool, error) {
	return u.libraryRepo.IsFavorite(ctx, userID, mediaType, providerMediaID)
}

// Custom Collections
func (u *LibraryUsecase) CreateCustomList(ctx context.Context, list *domain.CustomList) error {
	list.Title = strings.TrimSpace(list.Title)
	if list.Title == "" {
		return domain.NewAppError(domain.ErrCodeValidation, "list title cannot be empty", nil)
	}
	return u.libraryRepo.CreateCustomList(ctx, list)
}

func (u *LibraryUsecase) GetCustomLists(ctx context.Context, userID string) ([]domain.CustomList, error) {
	return u.libraryRepo.GetCustomLists(ctx, userID)
}

func (u *LibraryUsecase) GetCustomListDetails(ctx context.Context, listID, userID string) (*domain.CustomList, []domain.CustomListItem, error) {
	return u.libraryRepo.GetCustomListByID(ctx, listID, userID)
}

func (u *LibraryUsecase) DeleteCustomList(ctx context.Context, listID, userID string) error {
	return u.libraryRepo.DeleteCustomList(ctx, listID, userID)
}

func (u *LibraryUsecase) AddCustomListItem(ctx context.Context, userID string, item *domain.CustomListItem) error {
	if item.ListID == "" || item.ProviderMediaID == 0 {
		return domain.NewAppError(domain.ErrCodeValidation, "invalid item parameters", nil)
	}
	return u.libraryRepo.AddCustomListItem(ctx, userID, item)
}

func (u *LibraryUsecase) RemoveCustomListItem(ctx context.Context, userID, itemID string) error {
	return u.libraryRepo.RemoveCustomListItem(ctx, userID, itemID)
}

func (u *LibraryUsecase) UpdateWatchStatus(ctx context.Context, userID, itemID string, status domain.WatchStatus) error {
	return u.libraryRepo.UpdateWatchStatus(ctx, userID, itemID, status)
}
