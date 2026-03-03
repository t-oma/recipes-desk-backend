package recipe

import (
	"fmt"
	"strings"
)

const (
	UnitCup  = "cup"
	UnitGram = "g"
	UnitPcs  = "pcs"
)

var availableUnits = map[string]bool{ //nolint:gochecknoglobals // constant
	UnitCup:  true,
	UnitGram: true,
	UnitPcs:  true,
}

type Unit struct {
	name string
}

func NewUnit(name string) (Unit, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Unit{}, ErrUnitEmptyName
	}
	if !availableUnits[name] {
		return Unit{}, ErrUnitUnknown
	}

	return Unit{
		name: name,
	}, nil
}

func (u Unit) Name() string {
	return u.name
}

func (u Unit) String() string {
	return u.name
}

var (
	ErrUnitEmptyName = fmt.Errorf(
		"%w: unit name cannot be empty",
		ErrValidation,
	)
	ErrUnitUnknown = fmt.Errorf(
		"%w: unit is unknown",
		ErrValidation,
	)
)
