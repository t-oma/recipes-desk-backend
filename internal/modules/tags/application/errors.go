package application

import (
	"errors"
	"fmt"

	"recipes-desk/internal/modules/tags/domain"
)

var (
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = domain.ErrNotFound
	// ErrValidation indicates a validation error.
	ErrValidation = domain.ErrValidation
	// ErrConflict indicates a resource conflict.
	ErrConflict = domain.ErrConflict
	// ErrTimeout indicates a database operation timeout.
	ErrTimeout = domain.ErrTimeout
	// ErrInternal indicates an internal server error.
	ErrInternal = errors.New("internal server error")
	// ErrTagAlreadyExists indicates a tag already exists.
	ErrTagAlreadyExists = fmt.Errorf("%w: tag already exists", domain.ErrConflict)
	// ErrServiceUnavailable indicates a service unavailable error.
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
)
