package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/auth/domain"
)

const (
	LastNameMinLength = 2
	LastNameMaxLength = 50
)

type LastName struct {
	value string
}

func NewLastName(lastName string) (LastName, error) {
	lastName = strings.TrimSpace(lastName)
	if lastName == "" {
		return LastName{}, ErrLastNameEmpty
	}
	if len(lastName) < LastNameMinLength {
		return LastName{}, ErrLastNameTooShort
	}
	if len(lastName) > LastNameMaxLength {
		return LastName{}, ErrLastNameTooLong
	}

	return LastName{
		value: lastName,
	}, nil
}

func (l LastName) String() string {
	return l.value
}

var (
	ErrLastNameEmpty = fmt.Errorf(
		"%w: last name cannot be empty",
		domain.ErrValidation,
	)
	ErrLastNameTooShort = fmt.Errorf(
		"%w: last name must be at least %d characters",
		domain.ErrValidation,
		LastNameMinLength,
	)
	ErrLastNameTooLong = fmt.Errorf(
		"%w: last name must be at most %d characters",
		domain.ErrValidation,
		LastNameMaxLength,
	)
)
