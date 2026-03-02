package domain

import (
	"errors"
	"fmt"
	"time"
)

// Recipe represents a cooking recipe.
type Recipe struct {
	ID          string
	Title       string
	Description string
	Ingredients []Ingredient
	Steps       []Step
	CookingTime int
	Portions    int
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Ingredient represents a recipe ingredient.
type Ingredient struct {
	Name   string
	Amount float64
	Unit   string
}

// Step represents a cooking step.
type Step struct {
	Order       int
	Description string
	Duration    int // in minutes
}

// Validate performs business validation on the recipe.
func (r *Recipe) Validate() error {
	if r.Title == "" {
		return ErrEmptyTitle
	}
	if len(r.Title) < 3 || len(r.Title) > 200 {
		return ErrInvalidTitleLength
	}
	if r.Description == "" {
		return ErrEmptyDescription
	}
	if len(r.Description) < 10 || len(r.Description) > 5000 {
		return ErrInvalidDescriptionLength
	}
	if len(r.Ingredients) == 0 {
		return ErrNoIngredients
	}
	if len(r.Steps) == 0 {
		return ErrNoSteps
	}
	if r.CookingTime <= 0 {
		return ErrInvalidCookingTime
	}
	if r.Portions < 1 || r.Portions > 100 {
		return ErrInvalidPortions
	}
	if len(r.Tags) == 0 {
		return ErrNoTags
	}
	return nil
}

var ErrValidation = errors.New("validation error")

// Validation errors - wrapped with ErrValidation category.
var (
	ErrEmptyTitle         = fmt.Errorf("%w: title cannot be empty", ErrValidation)
	ErrInvalidTitleLength = fmt.Errorf(
		"%w: title must be between 3 and 200 characters",
		ErrValidation,
	)
	ErrEmptyDescription         = fmt.Errorf("%w: description cannot be empty", ErrValidation)
	ErrInvalidDescriptionLength = fmt.Errorf(
		"%w: description must be between 10 and 5000 characters",
		ErrValidation,
	)
	ErrNoIngredients = fmt.Errorf(
		"%w: recipe must have at least one ingredient",
		ErrValidation,
	)
	ErrNoSteps = fmt.Errorf(
		"%w: recipe must have at least one step",
		ErrValidation,
	)
	ErrInvalidCookingTime = fmt.Errorf(
		"%w: cooking time must be greater than 0",
		ErrValidation,
	)
	ErrInvalidPortions = fmt.Errorf(
		"%w: portions must be between 1 and 100",
		ErrValidation,
	)
	ErrNoTags = fmt.Errorf("%w: recipe must have at least one tag", ErrValidation)
)
