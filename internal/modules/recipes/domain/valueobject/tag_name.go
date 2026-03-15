package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/recipes/domain"
)

const (
	TagNameMaxLength = 50
)

type TagName struct {
	value string
}

func NewTagName(name string) (TagName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return TagName{}, ErrTagEmptyName
	}
	if len(name) > TagNameMaxLength {
		return TagName{}, ErrTagNameTooLong
	}
	return TagName{
		value: name,
	}, nil
}

func (t TagName) String() string {
	return t.value
}

var (
	ErrTagEmptyName = fmt.Errorf(
		"%w: tag name cannot be empty",
		domain.ErrValidation,
	)
	ErrTagNameTooLong = fmt.Errorf(
		"%w: tag name must be at most %d characters",
		domain.ErrValidation,
		TagNameMaxLength,
	)
)
