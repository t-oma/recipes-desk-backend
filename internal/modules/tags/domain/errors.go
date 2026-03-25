// Package domain contains tag domain errors.
package domain

import "errors"

var (
	// ErrValidation indicates a validation error.
	ErrValidation = errors.New("validation error")
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = errors.New("not found")
	// ErrForbidden indicates the operation is forbidden.
	ErrForbidden = errors.New("forbidden")
	// ErrConflict indicates a resource conflict (e.g., duplicate key).
	ErrConflict = errors.New("resource conflict")
	// ErrTimeout indicates a database operation timeout.
	ErrTimeout = errors.New("database timeout")
	// ErrDatabase indicates a generic database error.
	ErrDatabase = errors.New("database error")
	// ErrInternal indicates an internal server error.
	ErrInternal = errors.New("internal server error")
)
