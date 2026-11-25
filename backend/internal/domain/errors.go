package domain

import (
	"errors"
	"fmt"
)

// Standard domain error codes
const (
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeRateLimited      = "RATE_LIMITED"
	ErrCodeExternalProvider = "EXTERNAL_PROVIDER_ERROR"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeBadRequest       = "BAD_REQUEST"
)

// AppError represents an application error with a machine-readable code, message, and optional details.
type AppError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Err     error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Sentinel domain errors
var (
	ErrNotFound       = NewAppError(ErrCodeNotFound, "resource not found", nil)
	ErrUnauthorized   = NewAppError(ErrCodeUnauthorized, "unauthorized access", nil)
	ErrForbidden      = NewAppError(ErrCodeForbidden, "forbidden action", nil)
	ErrInternal       = NewAppError(ErrCodeInternal, "an unexpected internal error occurred", nil)
	ErrRateLimited    = NewAppError(ErrCodeRateLimited, "rate limit exceeded, please try again later", nil)
	ErrProviderFailed = NewAppError(ErrCodeExternalProvider, "upstream media provider request failed", nil)
)

// NewAppError creates a new AppError.
func NewAppError(code, message string, details interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WrapError wraps an existing error into an AppError.
func WrapError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsAppError checks if the given error matches an AppError code.
func IsAppError(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}
