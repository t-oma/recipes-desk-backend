package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/auth/domain"
)

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return Email{}, ErrEmailEmpty
	}

	return Email{
		value: email,
	}, nil
}

func (e Email) String() string {
	return e.value
}

var ErrEmailEmpty = fmt.Errorf(
	"%w: email cannot be empty",
	domain.ErrValidation,
)
