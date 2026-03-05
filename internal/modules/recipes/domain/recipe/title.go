package recipe

import (
	"fmt"
	"strings"
)

const (
	TitleMinLength = 4
	TitleMaxLength = 200
)

type Title struct {
	value string
}

func NewTitle(title string) (Title, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return Title{}, ErrTitleEmpty
	}
	if len(title) < TitleMinLength {
		return Title{}, ErrTitleTooShort
	}
	if len(title) > TitleMaxLength {
		return Title{}, ErrTitleTooLong
	}
	return Title{
		value: title,
	}, nil
}

func (t Title) String() string {
	return t.value
}

var (
	ErrTitleEmpty = fmt.Errorf(
		"%w: recipe title cannot be empty",
		ErrValidation,
	)
	ErrTitleTooShort = fmt.Errorf(
		"%w: recipe title must be at least %d characters",
		ErrValidation,
		TitleMinLength,
	)
	ErrTitleTooLong = fmt.Errorf(
		"%w: recipe title must be at most %d characters",
		ErrValidation,
		TitleMaxLength,
	)
)
