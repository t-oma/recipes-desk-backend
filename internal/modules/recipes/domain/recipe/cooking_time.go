package recipe

import (
	"fmt"
	"time"
)

type CookingTime struct {
	value time.Duration
}

func NewCookingTime(seconds int) (CookingTime, error) {
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

func (c CookingTime) String() string {
	return fmt.Sprintf("%f sec", c.value.Seconds())
}

var ErrCookingNegativeTime = fmt.Errorf(
	"%w: cooking time cannot be negative",
	ErrValidation,
)
