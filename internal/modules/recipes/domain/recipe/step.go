package recipe

import (
	"fmt"
	"time"
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

func NewStep(order int, description string, durationMin int) (Step, error) {
	if order < 1 {
		return Step{}, ErrInvalidStepOrder
	}
	if len(description) < StepMinDescriptionLength {
		return Step{}, ErrStepDescriptionTooShort
	}
	if len(description) > StepMaxDescriptionLength {
		return Step{}, ErrStepDescriptionTooLong
	}
	if durationMin < 0 {
		return Step{}, ErrNegativeDuration
	}

	return Step{
		order:       order,
		description: description,
		duration:    time.Duration(durationMin) * time.Minute,
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

func (s Step) String() string {
	return fmt.Sprintf("%d. %s (%d min)", s.order, s.description, s.duration)
}

var (
	ErrInvalidStepOrder = fmt.Errorf(
		"%w: step order must be greater than 0",
		ErrValidation,
	)
	ErrNegativeDuration = fmt.Errorf(
		"%w: step duration cannot be negative",
		ErrValidation,
	)
	ErrStepDescriptionTooShort = fmt.Errorf(
		"%w: step description must be at least %d characters",
		ErrValidation,
		StepMinDescriptionLength,
	)
	ErrStepDescriptionTooLong = fmt.Errorf(
		"%w: step description must be at most %d characters",
		ErrValidation,
		StepMaxDescriptionLength,
	)
)
