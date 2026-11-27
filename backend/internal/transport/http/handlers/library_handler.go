package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/middleware"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/response"
	"github.com/Amirh2006/Oryn/backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

type LibraryHandler struct {
	usecase *usecase.LibraryUsecase
}

func NewLibraryHandler(u *usecase.LibraryUsecase) *LibraryHandler {
	return &LibraryHandler{usecase: u}
}

// User Preferences Endpoints

// GetPreferences handles GET /api/v1/users/preferences
func (h *LibraryHandler) GetPreferences(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	prefs, err := h.usecase.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, prefs)
}

// UpdatePreferences handles PUT /api/v1/users/preferences
func (h *LibraryHandler) UpdatePreferences(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var prefs domain.UserPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid preferences payload", err.Error())
		return
	}

	prefs.UserID = userID
	if err := h.usecase.UpdatePreferences(c.Request.Context(), &prefs); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, prefs)
}

// Favorites Endpoints

type AddFavoriteRequest struct {
	ProviderName    string  `json:"providerName"`
	ProviderMediaID int     `json:"providerMediaId" binding:"required"`
	MediaType       string  `json:"mediaType" binding:"required"`
	Title           string  `json:"title" binding:"required"`
	PosterURL       string  `json:"posterUrl"`
	Rating          float64 `json:"rating"`
	ReleaseYear     int     `json:"releaseYear"`
}

// GetFavorites handles GET /api/v1/library/favorites
func (h *LibraryHandler) GetFavorites(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, total, err := h.usecase.GetFavorites(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	response.Paginated(c, http.StatusOK, items, response.PaginationMeta{
		Page:       page,
		PageSize:   len(items),
		TotalItems: total,
		TotalPages: totalPages,
	})
}

// AddFavorite handles POST /api/v1/library/favorites
func (h *LibraryHandler) AddFavorite(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var req AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid favorite payload", err.Error())
		return
	}

	providerName := req.ProviderName
	if providerName == "" {
		providerName = "tmdb"
	}

	fav := &domain.FavoriteItem{
		UserID:          userID,
		ProviderName:    providerName,
		ProviderMediaID: req.ProviderMediaID,
		MediaType:       req.MediaType,
		Title:           req.Title,
		PosterURL:       req.PosterURL,
		Rating:          req.Rating,
		ReleaseYear:     req.ReleaseYear,
	}

	if err := h.usecase.AddFavorite(c.Request.Context(), fav); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, fav)
}

// RemoveFavorite handles DELETE /api/v1/library/favorites/:type/:id
func (h *LibraryHandler) RemoveFavorite(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	mediaType := c.Param("type")
	providerMediaID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid media id", nil)
		return
	}

	if err := h.usecase.RemoveFavorite(c.Request.Context(), userID, mediaType, providerMediaID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"message": "favorite removed successfully"})
}

// CheckFavorite handles GET /api/v1/library/favorites/:type/:id/check
func (h *LibraryHandler) CheckFavorite(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	mediaType := c.Param("type")
	providerMediaID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid media id", nil)
		return
	}

	isFav, err := h.usecase.IsFavorite(c.Request.Context(), userID, mediaType, providerMediaID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"isFavorite": isFav})
}

// Custom Collections Endpoints

type CreateCollectionRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic"`
}

// GetCollections handles GET /api/v1/library/collections
func (h *LibraryHandler) GetCollections(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	lists, err := h.usecase.GetCustomLists(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, lists)
}

// CreateCollection handles POST /api/v1/library/collections
func (h *LibraryHandler) CreateCollection(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid collection payload", err.Error())
		return
	}

	list := &domain.CustomList{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}

	if err := h.usecase.CreateCustomList(c.Request.Context(), list); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, list)
}

// GetCollectionDetails handles GET /api/v1/library/collections/:id
func (h *LibraryHandler) GetCollectionDetails(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	listID := c.Param("id")

	list, items, err := h.usecase.GetCustomListDetails(c.Request.Context(), listID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"collection": list,
		"items":      items,
	})
}

// DeleteCollection handles DELETE /api/v1/library/collections/:id
func (h *LibraryHandler) DeleteCollection(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	listID := c.Param("id")

	if err := h.usecase.DeleteCustomList(c.Request.Context(), listID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"message": "collection deleted successfully"})
}

type AddCollectionItemRequest struct {
	ProviderName    string             `json:"providerName"`
	ProviderMediaID int                `json:"providerMediaId" binding:"required"`
	MediaType       string             `json:"mediaType" binding:"required"`
	Title           string             `json:"title" binding:"required"`
	PosterURL       string             `json:"posterUrl"`
	WatchStatus     domain.WatchStatus `json:"watchStatus"`
}

// AddCollectionItem handles POST /api/v1/library/collections/:id/items
func (h *LibraryHandler) AddCollectionItem(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	listID := c.Param("id")

	var req AddCollectionItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid item payload", err.Error())
		return
	}

	providerName := req.ProviderName
	if providerName == "" {
		providerName = "tmdb"
	}
	status := req.WatchStatus
	if status == "" {
		status = domain.WatchStatusPlanToWatch
	}

	item := &domain.CustomListItem{
		ListID:          listID,
		ProviderName:    providerName,
		ProviderMediaID: req.ProviderMediaID,
		MediaType:       req.MediaType,
		Title:           req.Title,
		PosterURL:       req.PosterURL,
		WatchStatus:     status,
	}

	if err := h.usecase.AddCustomListItem(c.Request.Context(), userID, item); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, item)
}

// RemoveCollectionItem handles DELETE /api/v1/library/collections/:id/items/:itemId
func (h *LibraryHandler) RemoveCollectionItem(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	itemID := c.Param("itemId")

	if err := h.usecase.RemoveCustomListItem(c.Request.Context(), userID, itemID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"message": "item removed from collection"})
}

type UpdateWatchStatusRequest struct {
	Status domain.WatchStatus `json:"status" binding:"required"`
}

// UpdateWatchStatus handles PATCH /api/v1/library/collections/:id/items/:itemId/status
func (h *LibraryHandler) UpdateWatchStatus(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	itemID := c.Param("itemId")

	var req UpdateWatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, domain.ErrCodeValidation, "invalid status payload", err.Error())
		return
	}

	if err := h.usecase.UpdateWatchStatus(c.Request.Context(), userID, itemID, req.Status); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"message": "watch status updated successfully",
		"status":  req.Status,
	})
}
