// Package valueobject contains tag value object definitions.
package valueobject

import (
	"fmt"

	"recipes-desk/internal/modules/tags/domain"
)

type TagID struct {
	value string
}

func NewTagID(id string) (TagID, error) {
	if id == "" {
		return TagID{}, ErrTagIDEmpty
	}
	return TagID{value: id}, nil
}

func (t TagID) String() string {
	return t.value
}

var ErrTagIDEmpty = fmt.Errorf(
	"%w: tag ID cannot be empty",
	domain.ErrValidation,
)
