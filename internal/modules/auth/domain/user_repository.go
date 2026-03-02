package domain //nolint:cyclop // TODO: refactor

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
	ErrValidation    = errors.New("validation error")
)

var (
	ErrEmptyEmail         = fmt.Errorf("%w: email cannot be empty", ErrValidation)
	ErrInvalidEmail       = fmt.Errorf("%w: email is invalid", ErrValidation)
	ErrInvalidEmailLength = fmt.Errorf(
		"%w: email must be between 3 and 254 characters",
		ErrValidation,
	)
	ErrEmptyPassword = fmt.Errorf(
		"%w: password cannot be empty",
		ErrValidation,
	)
	ErrPasswordTooShort = fmt.Errorf(
		"%w: password must be at least 8 characters",
		ErrValidation,
	)
	ErrPasswordTooLong = fmt.Errorf(
		"%w: password must be at most 72 characters",
		ErrValidation,
	)
	ErrInvalidCredentials = fmt.Errorf(
		"%w: invalid credentials",
		ErrValidation,
	)
	ErrEmptyFirstName = fmt.Errorf(
		"%w: first name cannot be empty",
		ErrValidation,
	)
	ErrInvalidFirstNameLength = fmt.Errorf(
		"%w: first name must be between 2 and 50 characters",
		ErrValidation,
	)
	ErrEmptyLastName = fmt.Errorf(
		"%w: last name cannot be empty",
		ErrValidation,
	)
	ErrInvalidLastNameLength = fmt.Errorf(
		"%w: last name must be between 2 and 50 characters",
		ErrValidation,
	)
)

type UsersRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
