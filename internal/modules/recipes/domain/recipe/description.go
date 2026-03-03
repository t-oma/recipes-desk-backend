package recipe

import (
	"fmt"
	"strings"
)

const (
	DescriptionMinLength = 10
	DescriptionMaxLength = 5000
)

type Description struct {
	value string
}

func NewDescription(desc string) (Description, error) {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return Description{}, ErrDescriptionEmpty
	}
	if len(desc) < DescriptionMinLength {
		return Description{}, ErrDescriptionTooShort
	}
	if len(desc) > DescriptionMaxLength {
		return Description{}, ErrDescriptionTooLong
	}

	return Description{
		value: desc,
	}, nil
}

func (d Description) Value() string {
	return d.value
}

func (d Description) String() string {
	return d.value
}

var (
	ErrDescriptionEmpty    = fmt.Errorf("%w: description cannot be empty", ErrValidation)
	ErrDescriptionTooShort = fmt.Errorf(
		"%w: recipe description must be at least %d characters",
		ErrValidation,
		DescriptionMinLength,
	)

	ErrDescriptionTooLong = fmt.Errorf(
		"%w: recipe description must be at most %d characters",
		ErrValidation,
		DescriptionMaxLength,
	)
)
