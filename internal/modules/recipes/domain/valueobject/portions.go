package valueobject

import (
	"fmt"
	"strconv"

	"recipes-desk/internal/modules/recipes/domain"
)

const (
	PortionsMin = 1
	PortionsMax = 100
)

type Portions struct {
	value int
}

func NewPortions(portions int) (Portions, error) {
	if portions < PortionsMin {
		return Portions{}, ErrPortionsTooFew
	}
	if portions > PortionsMax {
		return Portions{}, ErrPortionsTooMany
	}
	return Portions{
		value: portions,
	}, nil
}

func (p Portions) Value() int {
	return p.value
}

func (p Portions) String() string {
	return strconv.Itoa(p.value)
}

var (
	ErrPortionsTooFew = fmt.Errorf(
		"%w: portions must be greater than %d",
		domain.ErrValidation,
		PortionsMin,
	)
	ErrPortionsTooMany = fmt.Errorf(
		"%w: portions must be at most %d",
		domain.ErrValidation,
		PortionsMax,
	)
)
