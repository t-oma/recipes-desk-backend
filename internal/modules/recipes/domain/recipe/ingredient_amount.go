package recipe

import "fmt"

const (
	AmountMin = 0
)

type Amount struct {
	value float64
}

func NewAmount(amount float64) (Amount, error) {
	if amount < AmountMin {
		return Amount{}, ErrAmountToFew
	}

	return Amount{
		value: amount,
	}, nil
}

func (a Amount) Value() float64 {
	return a.value
}

func (a Amount) String() string {
	return fmt.Sprintf("%f", a.value)
}

var ErrAmountToFew = fmt.Errorf(
	"%w: ingredient amount cannot be less than %d",
	ErrValidation,
	AmountMin,
)
