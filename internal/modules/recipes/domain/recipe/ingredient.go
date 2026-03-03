package recipe

import (
	"fmt"
	"strings"
)

type Ingredient struct {
	name   string
	amount Amount
	unit   Unit
}

func NewIngredient(name string, amount float64, unit string) (Ingredient, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Ingredient{}, ErrIngredientEmptyName
	}

	amountVO, err := NewAmount(amount)
	if err != nil {
		return Ingredient{}, err
	}

	unitVO, err := NewUnit(unit)
	if err != nil {
		return Ingredient{}, err
	}

	return Ingredient{
		name:   name,
		amount: amountVO,
		unit:   unitVO,
	}, nil
}

func (i Ingredient) Name() string {
	return i.name
}

func (i Ingredient) Amount() Amount {
	return i.amount
}

func (i Ingredient) Unit() Unit {
	return i.unit
}

func (i Ingredient) Equals(other *Ingredient) bool {
	if other == nil {
		return false
	}

	return i.name == other.name &&
		i.amount.Value() == other.amount.Value() &&
		i.unit.Name() == other.unit.Name()
}

func (i Ingredient) String() string {
	return fmt.Sprintf("%s %.2f %s", i.name, i.amount.value, i.unit.name)
}

var ErrIngredientEmptyName = fmt.Errorf(
	"%w: ingredient name cannot be empty",
	ErrValidation,
)
