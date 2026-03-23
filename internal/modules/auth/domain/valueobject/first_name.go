package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/auth/domain"
)

const (
	FirstNameMinLength = 2
	FirstNameMaxLength = 50
)

type FirstName struct {
	value string
}

func NewFirstName(firstName string) (FirstName, error) {
	firstName = strings.TrimSpace(firstName)
	if firstName == "" {
		return FirstName{}, ErrFirstNameEmpty
	}
	if len(firstName) < FirstNameMinLength {
		return FirstName{}, ErrFirstNameTooShort
	}
	if len(firstName) > FirstNameMaxLength {
		return FirstName{}, ErrFirstNameTooLong
	}

	return FirstName{
		value: firstName,
	}, nil
}

func (f FirstName) String() string {
	return f.value
}

var (
	ErrFirstNameEmpty = fmt.Errorf(
		"%w: first name cannot be empty",
		domain.ErrValidation,
	)
	ErrFirstNameTooShort = fmt.Errorf(
		"%w: first name must be at least %d characters",
		domain.ErrValidation,
		FirstNameMinLength,
	)
	ErrFirstNameTooLong = fmt.Errorf(
		"%w: first name must be at most %d characters",
		domain.ErrValidation,
		FirstNameMaxLength,
	)
)
