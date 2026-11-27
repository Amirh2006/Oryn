package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/platform/security"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/middleware"
	"github.com/Amirh2006/Oryn/backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockLibraryRepo struct {
	favorites   map[string]domain.FavoriteItem
	lists       map[string]domain.CustomList
	listItems   map[string]domain.CustomListItem
}

func newMockLibraryRepo() *mockLibraryRepo {
	return &mockLibraryRepo{
		favorites: make(map[string]domain.FavoriteItem),
		lists:     make(map[string]domain.CustomList),
		listItems: make(map[string]domain.CustomListItem),
	}
}

func (m *mockLibraryRepo) AddFavorite(ctx context.Context, fav *domain.FavoriteItem) error {
	key := fav.UserID + ":" + fav.MediaType + ":" + string(rune(fav.ProviderMediaID))
	fav.ID = "fav-1"
	m.favorites[key] = *fav
	return nil
}

func (m *mockLibraryRepo) RemoveFavorite(ctx context.Context, userID, mediaType string, providerMediaID int) error {
	key := userID + ":" + mediaType + ":" + string(rune(providerMediaID))
	delete(m.favorites, key)
	return nil
}

func (m *mockLibraryRepo) GetFavorites(ctx context.Context, userID string, page, limit int) ([]domain.FavoriteItem, int64, error) {
	var list []domain.FavoriteItem
	for _, f := range m.favorites {
		if f.UserID == userID {
			list = append(list, f)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockLibraryRepo) IsFavorite(ctx context.Context, userID, mediaType string, providerMediaID int) (bool, error) {
	key := userID + ":" + mediaType + ":" + string(rune(providerMediaID))
	_, ok := m.favorites[key]
	return ok, nil
}

func (m *mockLibraryRepo) CreateCustomList(ctx context.Context, list *domain.CustomList) error {
	list.ID = "list-1"
	m.lists[list.ID] = *list
	return nil
}

func (m *mockLibraryRepo) GetCustomLists(ctx context.Context, userID string) ([]domain.CustomList, error) {
	var result []domain.CustomList
	for _, l := range m.lists {
		if l.UserID == userID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockLibraryRepo) GetCustomListByID(ctx context.Context, listID, userID string) (*domain.CustomList, []domain.CustomListItem, error) {
	l, ok := m.lists[listID]
	if !ok || l.UserID != userID {
		return nil, nil, domain.ErrNotFound
	}
	var items []domain.CustomListItem
	for _, item := range m.listItems {
		if item.ListID == listID {
			items = append(items, item)
		}
	}
	return &l, items, nil
}

func (m *mockLibraryRepo) DeleteCustomList(ctx context.Context, listID, userID string) error {
	l, ok := m.lists[listID]
	if !ok || l.UserID != userID {
		return domain.ErrNotFound
	}
	delete(m.lists, listID)
	return nil
}

func (m *mockLibraryRepo) AddCustomListItem(ctx context.Context, userID string, item *domain.CustomListItem) error {
	item.ID = "item-1"
	m.listItems[item.ID] = *item
	return nil
}

func (m *mockLibraryRepo) RemoveCustomListItem(ctx context.Context, userID, itemID string) error {
	delete(m.listItems, itemID)
	return nil
}

func (m *mockLibraryRepo) UpdateWatchStatus(ctx context.Context, userID, itemID string, status domain.WatchStatus) error {
	item, ok := m.listItems[itemID]
	if !ok {
		return domain.ErrNotFound
	}
	item.WatchStatus = status
	m.listItems[itemID] = item
	return nil
}

func setupLibraryTestRouter() (*gin.Engine, *LibraryHandler, *security.JWTManager, string) {
	gin.SetMode(gin.TestMode)
	lRepo := newMockLibraryRepo()
	pRepo := &mockPrefRepo{prefs: make(map[string]*domain.UserPreferences)}

	jwtSecret := "super-secure-test-jwt-secret-key-32-bytes"
	jwtMgr := security.NewJWTManager(jwtSecret, 15*time.Minute, 24*time.Hour)

	usecase := usecase.NewLibraryUsecase(lRepo, pRepo)
	handler := NewLibraryHandler(usecase)

	tokenPair, _ := jwtMgr.GenerateTokenPair("user-library-1", "user@oryn.app")
	validToken := tokenPair.AccessToken

	r := gin.New()
	authMiddleware := middleware.Auth(jwtMgr)

	users := r.Group("/api/v1/users")
	users.Use(authMiddleware)
	{
		users.GET("/preferences", handler.GetPreferences)
		users.PUT("/preferences", handler.UpdatePreferences)
	}

	library := r.Group("/api/v1/library")
	library.Use(authMiddleware)
	{
		library.GET("/favorites", handler.GetFavorites)
		library.POST("/favorites", handler.AddFavorite)
		library.DELETE("/favorites/:type/:id", handler.RemoveFavorite)
		library.GET("/favorites/:type/:id/check", handler.CheckFavorite)

		library.GET("/collections", handler.GetCollections)
		library.POST("/collections", handler.CreateCollection)
		library.GET("/collections/:id", handler.GetCollectionDetails)
		library.DELETE("/collections/:id", handler.DeleteCollection)
		library.POST("/collections/:id/items", handler.AddCollectionItem)
		library.DELETE("/collections/:id/items/:itemId", handler.RemoveCollectionItem)
		library.PATCH("/collections/:id/items/:itemId/status", handler.UpdateWatchStatus)
	}

	return r, handler, jwtMgr, validToken
}

func TestLibraryHandler_Preferences(t *testing.T) {
	router, _, _, token := setupLibraryTestRouter()

	// 1. Update Preferences
	updatePayload := map[string]interface{}{
		"language":           "fa",
		"theme":              "light",
		"favoriteGenres":     []int{28, 878},
		"preferredRatingMin": 7.5,
		"maxRuntime":         120,
	}
	body, _ := json.Marshal(updatePayload)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/preferences", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. Get Preferences
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/preferences", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLibraryHandler_Favorites(t *testing.T) {
	router, _, _, token := setupLibraryTestRouter()

	// 1. Add Favorite
	favPayload := map[string]interface{}{
		"providerMediaId": 550,
		"mediaType":       "movie",
		"title":           "Fight Club",
		"rating":          8.8,
		"releaseYear":     1999,
	}
	body, _ := json.Marshal(favPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/library/favorites", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// 2. Check Favorite
	req = httptest.NewRequest(http.MethodGet, "/api/v1/library/favorites/movie/550/check", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// 3. Get Favorites
	req = httptest.NewRequest(http.MethodGet, "/api/v1/library/favorites?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// 4. Remove Favorite
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/library/favorites/movie/550", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLibraryHandler_Collections(t *testing.T) {
	router, _, _, token := setupLibraryTestRouter()

	// 1. Create Collection
	colPayload := map[string]interface{}{
		"title":       "Top Sci-Fi",
		"description": "My favorite sci-fi movies",
		"isPublic":    true,
	}
	body, _ := json.Marshal(colPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/library/collections", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// 2. Get Collections
	req = httptest.NewRequest(http.MethodGet, "/api/v1/library/collections", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// 3. Add Item to Collection
	itemPayload := map[string]interface{}{
		"providerMediaId": 157336,
		"mediaType":       "movie",
		"title":           "Interstellar",
		"watchStatus":     "plan_to_watch",
	}
	body, _ = json.Marshal(itemPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/library/collections/list-1/items", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// 4. Update Watch Status
	statusPayload := map[string]interface{}{
		"status": "watched",
	}
	body, _ = json.Marshal(statusPayload)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/library/collections/list-1/items/item-1/status", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// 5. Get Collection Details
	req = httptest.NewRequest(http.MethodGet, "/api/v1/library/collections/list-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// 6. Delete Collection
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/library/collections/list-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
