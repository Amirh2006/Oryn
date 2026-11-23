package response

import (
	"net/http"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

// Meta contains contextual request metadata.
type Meta struct {
	RequestID string      `json:"requestId,omitempty"`
	Timestamp string      `json:"timestamp"`
	Extra     interface{} `json:"extra,omitempty"`
}

// ErrorDetail describes an error with code and message.
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Response represents the unified API response envelope.
type Response struct {
	Data  interface{}  `json:"data"`
	Meta  Meta         `json:"meta"`
	Error *ErrorDetail `json:"error"`
}

// PaginationMeta provides pagination information.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

func buildMeta(c *gin.Context, extra interface{}) Meta {
	reqID := ""
	if c != nil {
		reqID = c.GetString("RequestID")
	}
	return Meta{
		RequestID: reqID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Extra:     extra,
	}
}

// JSON sends a successful response with HTTP status 200/201/etc.
func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Data:  data,
		Meta:  buildMeta(c, nil),
		Error: nil,
	})
}

// Paginated sends a paginated successful response.
func Paginated(c *gin.Context, status int, data interface{}, pagination PaginationMeta) {
	c.JSON(status, Response{
		Data:  data,
		Meta:  buildMeta(c, pagination),
		Error: nil,
	})
}

// Error sends an error response with appropriate envelope.
func Error(c *gin.Context, status int, code, message string, details interface{}) {
	c.JSON(status, Response{
		Data: nil,
		Meta: buildMeta(c, nil),
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// HandleError maps domain errors to appropriate HTTP responses.
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if appErr, ok := err.(*domain.AppError); ok {
		status := http.StatusInternalServerError
		switch appErr.Code {
		case domain.ErrCodeNotFound:
			status = http.StatusNotFound
		case domain.ErrCodeValidation, domain.ErrCodeBadRequest:
			status = http.StatusBadRequest
		case domain.ErrCodeUnauthorized:
			status = http.StatusUnauthorized
		case domain.ErrCodeForbidden:
			status = http.StatusForbidden
		case domain.ErrCodeConflict:
			status = http.StatusConflict
		case domain.ErrCodeRateLimited:
			status = http.StatusTooManyRequests
		case domain.ErrCodeExternalProvider:
			status = http.StatusBadGateway
		}
		Error(c, status, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	Error(c, http.StatusInternalServerError, domain.ErrCodeInternal, "an unexpected internal error occurred", nil)
}
