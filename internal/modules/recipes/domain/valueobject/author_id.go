package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/recipes/domain"
)

type AuthorID struct {
	value string
}

func NewAuthorID(id string) (AuthorID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return AuthorID{}, ErrAuthorIDEmpty
	}

	return AuthorID{
		value: id,
	}, nil
}

func (a AuthorID) String() string {
	return a.value
}

var ErrAuthorIDEmpty = fmt.Errorf(
	"%w: author ID cannot be empty",
	domain.ErrValidation,
)
