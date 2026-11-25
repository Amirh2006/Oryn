package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/repository/cache"
)

// CachedMediaProvider decorates any domain.MediaProvider with Redis caching
type CachedMediaProvider struct {
	provider domain.MediaProvider
	cache    *cache.RedisCache
}

// NewCachedMediaProvider constructs a new CachedMediaProvider
func NewCachedMediaProvider(provider domain.MediaProvider, c *cache.RedisCache) *CachedMediaProvider {
	return &CachedMediaProvider{
		provider: provider,
		cache:    c,
	}
}

func (c *CachedMediaProvider) GetTrending(ctx context.Context, mediaType domain.MediaType, timeWindow string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	cacheKey := fmt.Sprintf("media:trending:%s:%s:%d", mediaType, timeWindow, page)
	var cached domain.PaginatedResult[domain.MediaItem]
	if c.cache != nil && c.cache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	result, err := c.provider.GetTrending(ctx, mediaType, timeWindow, page)
	if err != nil {
		return nil, err
	}

	if c.cache != nil && result != nil {
		c.cache.Set(ctx, cacheKey, result, 1*time.Hour)
	}
	return result, nil
}

func (c *CachedMediaProvider) GetPopular(ctx context.Context, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	cacheKey := fmt.Sprintf("media:popular:%s:%d", mediaType, page)
	var cached domain.PaginatedResult[domain.MediaItem]
	if c.cache != nil && c.cache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	result, err := c.provider.GetPopular(ctx, mediaType, page)
	if err != nil {
		return nil, err
	}

	if c.cache != nil && result != nil {
		c.cache.Set(ctx, cacheKey, result, 2*time.Hour)
	}
	return result, nil
}

func (c *CachedMediaProvider) GetTopRated(ctx context.Context, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	cacheKey := fmt.Sprintf("media:top_rated:%s:%d", mediaType, page)
	var cached domain.PaginatedResult[domain.MediaItem]
	if c.cache != nil && c.cache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	result, err := c.provider.GetTopRated(ctx, mediaType, page)
	if err != nil {
		return nil, err
	}

	if c.cache != nil && result != nil {
		c.cache.Set(ctx, cacheKey, result, 6*time.Hour)
	}
	return result, nil
}

func (c *CachedMediaProvider) SearchMedia(ctx context.Context, query string, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	cacheKey := fmt.Sprintf("media:search:%s:%s:%d", mediaType, strings.ToLower(strings.TrimSpace(query)), page)
	var cached domain.PaginatedResult[domain.MediaItem]
	if c.cache != nil && c.cache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	result, err := c.provider.SearchMedia(ctx, query, mediaType, page)
	if err != nil {
		return nil, err
	}

	if c.cache != nil && result != nil {
		c.cache.Set(ctx, cacheKey, result, 15*time.Minute)
	}
	return result, nil
}

func (c *CachedMediaProvider) DiscoverMedia(ctx context.Context, filter domain.MediaFilter, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	cacheKey := fmt.Sprintf("media:discover:%s:%v:%d:%.1f:%.1f:%s:%d",
		filter.MediaType, filter.Genres, filter.ReleaseYear, filter.RatingMin, filter.RatingMax, filter.SortBy, page)
	var cached domain.PaginatedResult[domain.MediaItem]
	if c.cache != nil && c.cache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	result, err := c.provider.DiscoverMedia(ctx, filter, page)
	if err != nil {
		return nil, err
	}

	if c.cache != nil && result != nil {
		c.cache.Set(ctx, cacheKey, result, 30*time.Minute)
	}
	return result, nil
}

func (c *CachedMediaProvider) GetMediaDetails(ctx context.Context, mediaType domain.MediaType, id string) (*domain.MediaDetails, error) {
	cacheKey := fmt.Sprintf("media:details:%s:%s", mediaType, id)
	var cached domain.MediaDetails
	if c.cache != nil && c.cache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	result, err := c.provider.GetMediaDetails(ctx, mediaType, id)
	if err != nil {
		return nil, err
	}

	if c.cache != nil && result != nil {
		c.cache.Set(ctx, cacheKey, result, 12*time.Hour)
	}
	return result, nil
}

func (c *CachedMediaProvider) GetSimilarMedia(ctx context.Context, mediaType domain.MediaType, id string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	cacheKey := fmt.Sprintf("media:similar:%s:%s:%d", mediaType, id, page)
	var cached domain.PaginatedResult[domain.MediaItem]
	if c.cache != nil && c.cache.Get(ctx, cacheKey, &cached) {
		return &cached, nil
	}

	result, err := c.provider.GetSimilarMedia(ctx, mediaType, id, page)
	if err != nil {
		return nil, err
	}

	if c.cache != nil && result != nil {
		c.cache.Set(ctx, cacheKey, result, 4*time.Hour)
	}
	return result, nil
}

func (c *CachedMediaProvider) GetGenres(ctx context.Context, mediaType domain.MediaType) ([]domain.Genre, error) {
	cacheKey := fmt.Sprintf("media:genres:%s", mediaType)
	var cached []domain.Genre
	if c.cache != nil && c.cache.Get(ctx, cacheKey, &cached) {
		return cached, nil
	}

	result, err := c.provider.GetGenres(ctx, mediaType)
	if err != nil {
		return nil, err
	}

	if c.cache != nil && len(result) > 0 {
		c.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	}
	return result, nil
}
