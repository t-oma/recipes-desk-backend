package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/auth/domain"
)

type RefreshTokenID struct {
	value string
}

func NewRefreshTokenID(id string) (RefreshTokenID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return RefreshTokenID{}, ErrRefreshTokenIDEmpty
	}

	return RefreshTokenID{
		value: id,
	}, nil
}

func (r RefreshTokenID) String() string {
	return r.value
}

var ErrRefreshTokenIDEmpty = fmt.Errorf(
	"%w: refresh token ID cannot be empty",
	domain.ErrValidation,
)
