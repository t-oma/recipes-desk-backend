package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/auth/domain"
)

type UserID struct {
	value string
}

func NewUserID(id string) (UserID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return UserID{}, ErrUserIDEmpty
	}

	return UserID{
		value: id,
	}, nil
}

func (u UserID) String() string {
	return u.value
}

var ErrUserIDEmpty = fmt.Errorf(
	"%w: user ID cannot be empty",
	domain.ErrValidation,
)
