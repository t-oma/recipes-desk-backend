package domain

import "errors"

var (
	ErrValidation = errors.New("validation error")
	ErrForbidden  = errors.New("forbidden")
)
