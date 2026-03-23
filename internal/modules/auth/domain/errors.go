package domain

import (
	"errors"
	"fmt"
)

var (
	ErrValidation   = errors.New("validation error")
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("resource conflict")
	ErrUnauthorized = errors.New("unauthorized")

	ErrUserAlreadyExists  = fmt.Errorf("%w: user already exists", ErrConflict)
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrTokenInvalid  = fmt.Errorf("%w: invalid token", ErrUnauthorized)
	ErrTokenExpired  = fmt.Errorf("%w: token has expired", ErrUnauthorized)
	ErrTokenNotFound = fmt.Errorf("%w: refresh token not found", ErrUnauthorized)

	ErrUserNotFound = fmt.Errorf("%w: user not found", ErrNotFound)

	// ErrTimeout indicates a database operation timeout.
	ErrTimeout = errors.New("database timeout")
	// ErrDatabase indicates a generic database error.
	ErrDatabase = errors.New("database error")
)
