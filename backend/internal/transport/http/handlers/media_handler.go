package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/transport/http/response"
	"github.com/Amirh2006/Oryn/backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

// MediaHandler handles incoming HTTP requests for media catalog exploration
type MediaHandler struct {
	usecase *usecase.MediaUsecase
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler(u *usecase.MediaUsecase) *MediaHandler {
	return &MediaHandler{usecase: u}
}

// GetTrending handles GET /api/v1/media/trending
func (h *MediaHandler) GetTrending(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "movie")
	timeWindow := c.DefaultQuery("timeWindow", "week")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.usecase.GetTrending(c.Request.Context(), mediaType, timeWindow, page)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, result.Items, response.PaginationMeta{
		Page:       result.Page,
		PageSize:   len(result.Items),
		TotalItems: result.TotalResults,
		TotalPages: result.TotalPages,
	})
}

// GetPopular handles GET /api/v1/media/popular
func (h *MediaHandler) GetPopular(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "movie")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.usecase.GetPopular(c.Request.Context(), mediaType, page)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, result.Items, response.PaginationMeta{
		Page:       result.Page,
		PageSize:   len(result.Items),
		TotalItems: result.TotalResults,
		TotalPages: result.TotalPages,
	})
}

// GetTopRated handles GET /api/v1/media/top-rated
func (h *MediaHandler) GetTopRated(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "movie")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.usecase.GetTopRated(c.Request.Context(), mediaType, page)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, result.Items, response.PaginationMeta{
		Page:       result.Page,
		PageSize:   len(result.Items),
		TotalItems: result.TotalResults,
		TotalPages: result.TotalPages,
	})
}

// Search handles GET /api/v1/media/search
func (h *MediaHandler) Search(c *gin.Context) {
	query := c.Query("q")
	mediaType := c.DefaultQuery("type", "movie")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.usecase.Search(c.Request.Context(), query, mediaType, page)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, result.Items, response.PaginationMeta{
		Page:       result.Page,
		PageSize:   len(result.Items),
		TotalItems: result.TotalResults,
		TotalPages: result.TotalPages,
	})
}

// Discover handles GET /api/v1/media/discover
func (h *MediaHandler) Discover(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "movie")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	year, _ := strconv.Atoi(c.Query("year"))
	ratingMin, _ := strconv.ParseFloat(c.Query("ratingMin"), 64)
	ratingMax, _ := strconv.ParseFloat(c.Query("ratingMax"), 64)
	runtimeMin, _ := strconv.Atoi(c.Query("runtimeMin"))
	runtimeMax, _ := strconv.Atoi(c.Query("runtimeMax"))
	sortBy := c.DefaultQuery("sortBy", "popularity.desc")
	lang := c.Query("lang")

	var genreIDs []int
	if genreStr := c.Query("genres"); genreStr != "" {
		for _, part := range strings.Split(genreStr, ",") {
			if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
				genreIDs = append(genreIDs, id)
			}
		}
	}

	filter := domain.MediaFilter{
		MediaType:        domain.MediaType(mediaType),
		Genres:           genreIDs,
		ReleaseYear:      year,
		RatingMin:        ratingMin,
		RatingMax:        ratingMax,
		RuntimeMin:       runtimeMin,
		RuntimeMax:       runtimeMax,
		OriginalLanguage: lang,
		SortBy:           sortBy,
	}

	result, err := h.usecase.Discover(c.Request.Context(), filter, page)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, result.Items, response.PaginationMeta{
		Page:       result.Page,
		PageSize:   len(result.Items),
		TotalItems: result.TotalResults,
		TotalPages: result.TotalPages,
	})
}

// GetDetails handles GET /api/v1/media/:type/:id
func (h *MediaHandler) GetDetails(c *gin.Context) {
	mediaType := c.Param("type")
	id := c.Param("id")

	details, err := h.usecase.GetDetails(c.Request.Context(), mediaType, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, details)
}

// GetSimilar handles GET /api/v1/media/:type/:id/similar
func (h *MediaHandler) GetSimilar(c *gin.Context) {
	mediaType := c.Param("type")
	id := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	result, err := h.usecase.GetSimilar(c.Request.Context(), mediaType, id, page)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, result.Items, response.PaginationMeta{
		Page:       result.Page,
		PageSize:   len(result.Items),
		TotalItems: result.TotalResults,
		TotalPages: result.TotalPages,
	})
}

// GetGenres handles GET /api/v1/media/genres
func (h *MediaHandler) GetGenres(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "movie")

	genres, err := h.usecase.GetGenres(c.Request.Context(), mediaType)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, genres)
}
