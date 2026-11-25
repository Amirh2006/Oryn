package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/response"
	"github.com/Amirh2006/Oryn/backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMediaProvider struct {
	trendingFn func(ctx context.Context, mediaType domain.MediaType, timeWindow string, page int) (*domain.PaginatedResult[domain.MediaItem], error)
	popularFn  func(ctx context.Context, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error)
	topRatedFn func(ctx context.Context, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error)
	searchFn   func(ctx context.Context, query string, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error)
	discoverFn func(ctx context.Context, filter domain.MediaFilter, page int) (*domain.PaginatedResult[domain.MediaItem], error)
	detailsFn  func(ctx context.Context, mediaType domain.MediaType, id string) (*domain.MediaDetails, error)
	similarFn  func(ctx context.Context, mediaType domain.MediaType, id string, page int) (*domain.PaginatedResult[domain.MediaItem], error)
	genresFn   func(ctx context.Context, mediaType domain.MediaType) ([]domain.Genre, error)
}

func (m *mockMediaProvider) GetTrending(ctx context.Context, mediaType domain.MediaType, timeWindow string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	return m.trendingFn(ctx, mediaType, timeWindow, page)
}
func (m *mockMediaProvider) GetPopular(ctx context.Context, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	return m.popularFn(ctx, mediaType, page)
}
func (m *mockMediaProvider) GetTopRated(ctx context.Context, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	return m.topRatedFn(ctx, mediaType, page)
}
func (m *mockMediaProvider) SearchMedia(ctx context.Context, query string, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	return m.searchFn(ctx, query, mediaType, page)
}
func (m *mockMediaProvider) DiscoverMedia(ctx context.Context, filter domain.MediaFilter, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	return m.discoverFn(ctx, filter, page)
}
func (m *mockMediaProvider) GetMediaDetails(ctx context.Context, mediaType domain.MediaType, id string) (*domain.MediaDetails, error) {
	return m.detailsFn(ctx, mediaType, id)
}
func (m *mockMediaProvider) GetSimilarMedia(ctx context.Context, mediaType domain.MediaType, id string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	return m.similarFn(ctx, mediaType, id, page)
}
func (m *mockMediaProvider) GetGenres(ctx context.Context, mediaType domain.MediaType) ([]domain.Genre, error) {
	return m.genresFn(ctx, mediaType)
}

func setupTestRouter(handler *MediaHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1/media")
	{
		api.GET("/trending", handler.GetTrending)
		api.GET("/popular", handler.GetPopular)
		api.GET("/top-rated", handler.GetTopRated)
		api.GET("/search", handler.Search)
		api.GET("/discover", handler.Discover)
		api.GET("/genres", handler.GetGenres)
		api.GET("/:type/:id", handler.GetDetails)
		api.GET("/:type/:id/similar", handler.GetSimilar)
	}
	return r
}

func TestMediaHandler_GetTrending(t *testing.T) {
	mockProv := &mockMediaProvider{
		trendingFn: func(ctx context.Context, mediaType domain.MediaType, timeWindow string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
			return &domain.PaginatedResult[domain.MediaItem]{
				Items: []domain.MediaItem{
					{ID: "movie-100", Title: "Oryn Movie", Rating: 8.5},
				},
				Page:         1,
				TotalPages:   10,
				TotalResults: 100,
			}, nil
		},
	}

	uc := usecase.NewMediaUsecase(mockProv)
	handler := NewMediaHandler(uc)
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/media/trending?type=movie&timeWindow=week&page=1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.Nil(t, res.Error)
	assert.NotNil(t, res.Data)
}

func TestMediaHandler_GetDetails_Success(t *testing.T) {
	mockProv := &mockMediaProvider{
		detailsFn: func(ctx context.Context, mediaType domain.MediaType, id string) (*domain.MediaDetails, error) {
			return &domain.MediaDetails{
				MediaItem: domain.MediaItem{
					ID:    "movie-550",
					Title: "Fight Club",
				},
				Tagline: "Mischief. Mayhem. Soap.",
				Runtime: 139,
			}, nil
		},
	}

	uc := usecase.NewMediaUsecase(mockProv)
	handler := NewMediaHandler(uc)
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/media/movie/550", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.Nil(t, res.Error)
}

func TestMediaHandler_Search_ValidationError(t *testing.T) {
	mockProv := &mockMediaProvider{}
	uc := usecase.NewMediaUsecase(mockProv)
	handler := NewMediaHandler(uc)
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/media/search?q=", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.NotNil(t, res.Error)
	assert.Equal(t, domain.ErrCodeValidation, res.Error.Code)
}
