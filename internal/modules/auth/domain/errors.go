package domain

import (
	"errors"
	"fmt"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
	ErrConflict   = errors.New("resource conflict")

	ErrUserAlreadyExists  = fmt.Errorf("%w: user already exists", ErrConflict)
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrTokenInvalid = errors.New("invalid token")
	ErrTokenExpired = errors.New("token has expired")

	ErrUserNotFound  = fmt.Errorf("%w: user not found", ErrNotFound)
	ErrTokenNotFound = fmt.Errorf("%w: refresh token not found", ErrNotFound)
)
