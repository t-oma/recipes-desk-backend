package domain

import "errors"

var (
	ErrValidation         = errors.New("validation error")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
