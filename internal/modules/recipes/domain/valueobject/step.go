package valueobject

import (
	"fmt"
	"time"

	"recipes-desk/internal/modules/recipes/domain"
)

const (
	StepMinDescriptionLength = 4
	StepMaxDescriptionLength = 1000
)

type Step struct {
	order       int
	description string
	duration    time.Duration // in minutes
}

func NewStep(order int, description string, durationSec int64) (Step, error) {
	if order < 1 {
		return Step{}, ErrStepInvalidOrder
	}
	if len(description) < StepMinDescriptionLength {
		return Step{}, ErrStepDescriptionTooShort
	}
	if len(description) > StepMaxDescriptionLength {
		return Step{}, ErrStepDescriptionTooLong
	}
	if durationSec < 0 {
		return Step{}, ErrStepNegativeDuration
	}

	return Step{
		order:       order,
		description: description,
		duration:    time.Duration(durationSec) * time.Second,
	}, nil
}

func (s Step) Order() int {
	return s.order
}

func (s Step) Description() string {
	return s.description
}

func (s Step) Duration() time.Duration {
	return s.duration
}

func (s Step) SecondsInt64() int64 {
	return int64(s.duration.Seconds())
}

func (s Step) String() string {
	return fmt.Sprintf("%d. %s (%d min)", s.order, s.description, s.duration)
}

var (
	ErrStepInvalidOrder = fmt.Errorf(
		"%w: step order must be greater than 0",
		domain.ErrValidation,
	)
	ErrStepNegativeDuration = fmt.Errorf(
		"%w: step duration cannot be negative",
		domain.ErrValidation,
	)
	ErrStepDescriptionTooShort = fmt.Errorf(
		"%w: step description must be at least %d characters",
		domain.ErrValidation,
		StepMinDescriptionLength,
	)
	ErrStepDescriptionTooLong = fmt.Errorf(
		"%w: step description must be at most %d characters",
		domain.ErrValidation,
		StepMaxDescriptionLength,
	)
)
