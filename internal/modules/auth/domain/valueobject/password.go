package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/auth/domain"
)

const (
	PasswordMinLength = 8
	PasswordMaxLength = 72
)

type Password struct {
	value string
}

func NewPassword(password string) (Password, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return Password{}, ErrPasswordEmpty
	}
	if len(password) < PasswordMinLength {
		return Password{}, ErrPasswordTooShort
	}
	if len(password) > PasswordMaxLength {
		return Password{}, ErrPasswordTooLong
	}

	return Password{
		value: password,
	}, nil
}

func (p Password) String() string {
	return p.value
}

var (
	ErrPasswordEmpty = fmt.Errorf(
		"%w: password cannot be empty",
		domain.ErrValidation,
	)
	ErrPasswordTooShort = fmt.Errorf(
		"%w: password must be at least %d characters",
		domain.ErrValidation,
		PasswordMinLength,
	)
	ErrPasswordTooLong = fmt.Errorf(
		"%w: password must be at most %d characters",
		domain.ErrValidation,
		PasswordMaxLength,
	)
)
