package application

import (
	"errors"

	"recipes-desk/internal/modules/auth/domain"
)

// Re-export domain errors for use by handlers.
// This provides a clean public API while keeping domain details encapsulated.
var (
	// ErrValidation indicates a validation error.
	ErrValidation = domain.ErrValidation
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = domain.ErrNotFound
	// ErrForbidden indicates the operation is forbidden.
	ErrForbidden = domain.ErrForbidden
	// ErrConflict indicates a resource conflict.
	ErrConflict = domain.ErrConflict
	// ErrUnauthorized indicates an unauthorized operation.
	ErrUnauthorized = domain.ErrUnauthorized

	// ErrUserNotFound indicates a user was not found.
	ErrUserNotFound = domain.ErrUserNotFound
	// ErrUserAlreadyExists indicates a user already exists.
	ErrUserAlreadyExists = domain.ErrUserAlreadyExists
	// ErrInvalidCredentials indicates invalid credentials.
	ErrInvalidCredentials = domain.ErrInvalidCredentials

	// ErrTokenNotFound indicates a refresh token was not found.
	ErrTokenNotFound = domain.ErrTokenNotFound
	// ErrTokenInvalid indicates an invalid token.
	ErrTokenInvalid = domain.ErrTokenInvalid
	// ErrTokenExpired indicates an expired token.
	ErrTokenExpired = domain.ErrTokenExpired

	// ErrTimeout indicates a database operation timeout.
	ErrTimeout = domain.ErrTimeout
	// ErrDatabase indicates a generic database error.
	ErrDatabase = domain.ErrDatabase

	// ErrServiceUnavailable indicates the service is temporarily unavailable.
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
	// ErrInternal indicates an internal server error.
	ErrInternal = errors.New("internal server error")
)
