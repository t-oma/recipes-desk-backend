// Package valueobject contains tag value object definitions.
package valueobject

import (
	"fmt"
	"strings"

	"recipes-desk/internal/modules/tags/domain"
)

const (
	TagNameMinLength = 2
	TagNameMaxLength = 50
)

type TagName struct {
	value string
}

func NewTagName(name string) (TagName, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return TagName{}, ErrTagNameEmpty
	}
	if len(name) < TagNameMinLength {
		return TagName{}, ErrTagNameTooShort
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
	ErrTagNameEmpty = fmt.Errorf(
		"%w: tag name cannot be empty",
		domain.ErrValidation,
	)
	ErrTagNameTooShort = fmt.Errorf(
		"%w: tag name must be at least %d characters",
		domain.ErrValidation,
		TagNameMinLength,
	)
	ErrTagNameTooLong = fmt.Errorf(
		"%w: tag name must be at most %d characters",
		domain.ErrValidation,
		TagNameMaxLength,
	)
)
