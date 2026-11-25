package usecase

import (
	"context"
	"strings"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
)

// MediaUsecase coordinates all media catalog operations
type MediaUsecase struct {
	provider domain.MediaProvider
}

// NewMediaUsecase creates a new MediaUsecase
func NewMediaUsecase(provider domain.MediaProvider) *MediaUsecase {
	return &MediaUsecase{provider: provider}
}

func (u *MediaUsecase) GetTrending(ctx context.Context, mediaType string, timeWindow string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	mType := normalizeMediaType(mediaType)
	if page < 1 {
		page = 1
	}
	return u.provider.GetTrending(ctx, mType, timeWindow, page)
}

func (u *MediaUsecase) GetPopular(ctx context.Context, mediaType string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	mType := normalizeMediaType(mediaType)
	if page < 1 {
		page = 1
	}
	return u.provider.GetPopular(ctx, mType, page)
}

func (u *MediaUsecase) GetTopRated(ctx context.Context, mediaType string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	mType := normalizeMediaType(mediaType)
	if page < 1 {
		page = 1
	}
	return u.provider.GetTopRated(ctx, mType, page)
}

func (u *MediaUsecase) Search(ctx context.Context, query string, mediaType string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, domain.NewAppError(domain.ErrCodeValidation, "search query cannot be empty", nil)
	}
	mType := normalizeMediaType(mediaType)
	if page < 1 {
		page = 1
	}
	return u.provider.SearchMedia(ctx, trimmedQuery, mType, page)
}

func (u *MediaUsecase) Discover(ctx context.Context, filter domain.MediaFilter, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	if filter.MediaType == "" {
		filter.MediaType = domain.MediaTypeMovie
	}
	if page < 1 {
		page = 1
	}
	return u.provider.DiscoverMedia(ctx, filter, page)
}

func (u *MediaUsecase) GetDetails(ctx context.Context, mediaType string, id string) (*domain.MediaDetails, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domain.NewAppError(domain.ErrCodeValidation, "media id is required", nil)
	}
	mType := normalizeMediaType(mediaType)
	return u.provider.GetMediaDetails(ctx, mType, id)
}

func (u *MediaUsecase) GetSimilar(ctx context.Context, mediaType string, id string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	if strings.TrimSpace(id) == "" {
		return nil, domain.NewAppError(domain.ErrCodeValidation, "media id is required", nil)
	}
	mType := normalizeMediaType(mediaType)
	if page < 1 {
		page = 1
	}
	return u.provider.GetSimilarMedia(ctx, mType, id, page)
}

func (u *MediaUsecase) GetGenres(ctx context.Context, mediaType string) ([]domain.Genre, error) {
	mType := normalizeMediaType(mediaType)
	return u.provider.GetGenres(ctx, mType)
}

func normalizeMediaType(m string) domain.MediaType {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case "tv", "series":
		return domain.MediaTypeTV
	case "movie", "film":
		return domain.MediaTypeMovie
	default:
		return domain.MediaTypeMovie
	}
}
