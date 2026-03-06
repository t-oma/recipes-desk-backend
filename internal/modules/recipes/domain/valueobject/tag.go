package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/recipes/domain"
)

const (
	TagMaxLength = 50
)

type Tag struct {
	name string
}

func NewTag(name string) (Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Tag{}, ErrTagEmptyName
	}
	if len(name) > TagMaxLength {
		return Tag{}, ErrTagNameTooLong
	}
	return Tag{
		name: name,
	}, nil
}

func (t Tag) Name() string {
	return t.name
}

func (t Tag) String() string {
	return t.name
}

var (
	ErrTagEmptyName = fmt.Errorf(
		"%w: tag name cannot be empty",
		domain.ErrValidation,
	)
	ErrTagNameTooLong = fmt.Errorf(
		"%w: tag name must be at most %d characters",
		domain.ErrValidation,
		TagMaxLength,
	)
)
