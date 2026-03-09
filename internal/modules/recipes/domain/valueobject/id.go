package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/recipes/domain"
)

type RecipeID struct {
	value string
}

func NewRecipeID(id string) (RecipeID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return RecipeID{}, ErrRecipeIDEmpty
	}

	return RecipeID{
		value: id,
	}, nil
}

func (r RecipeID) String() string {
	return r.value
}

var ErrRecipeIDEmpty = fmt.Errorf(
	"%w: recipe ID cannot be empty",
	domain.ErrValidation,
)
