package valueobject

import (
	"fmt"
	"time"

	"recipes-desk/internal/modules/recipes/domain"
)

type CookingTime struct {
	value time.Duration
}

func NewCookingTime(seconds int64) (CookingTime, error) {
	if seconds < 0 {
		return CookingTime{}, ErrCookingNegativeTime
	}
	return CookingTime{
		value: time.Duration(seconds) * time.Second,
	}, nil
}

func (c CookingTime) Duration() time.Duration {
	return c.value
}

func (c CookingTime) SecondsInt64() int64 {
	return int64(c.value.Seconds())
}

func (c CookingTime) String() string {
	return fmt.Sprintf("%f", c.value.Seconds())
}

var ErrCookingNegativeTime = fmt.Errorf(
	"%w: cooking time cannot be negative",
	domain.ErrValidation,
)
