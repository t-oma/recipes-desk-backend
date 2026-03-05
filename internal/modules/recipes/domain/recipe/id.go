package recipe

import (
	"fmt"
	"strings"
)

type EntityID struct {
	value string
}

func NewEntityID(id string) (EntityID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return EntityID{}, ErrRecipeIDEmpty
	}

	return EntityID{
		value: id,
	}, nil
}

func (r EntityID) String() string {
	return r.value
}

var ErrRecipeIDEmpty = fmt.Errorf(
	"%w: recipe ID cannot be empty",
	ErrValidation,
)
