package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrValidation indicates a validation error.
	ErrValidation = errors.New("validation error")
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = errors.New("not found")
	// ErrForbidden indicates the operation is forbidden.
	ErrForbidden = errors.New("forbidden")
	// ErrInternal indicates an internal server error.
	ErrInternal = errors.New("internal server error")

	// ErrRecipeNotFound indicates a recipe was not found.
	ErrRecipeNotFound = fmt.Errorf("%w: recipe not found", ErrNotFound)

	// ErrNoIngredients indicates that a recipe must have at least one ingredient.
	ErrNoIngredients = fmt.Errorf(
		"%w: recipe must have at least one ingredient",
		ErrValidation,
	)
	// ErrNoSteps indicates that a recipe must have at least one step.
	ErrNoSteps = fmt.Errorf("%w: recipe must have at least one step", ErrValidation)
	// ErrNoTags indicates that a recipe must have at least one tag.
	ErrNoTags = fmt.Errorf("%w: recipe must have at least one tag", ErrValidation)
	// ErrDuplicateIngredient indicates that a recipe has duplicate ingredients.
	ErrDuplicateIngredient = fmt.Errorf(
		"%w: recipe cannot have duplicate ingredients",
		ErrValidation,
	)
)

// ErrInvalidStepOrderf creates a validation error for invalid step order.
func ErrInvalidStepOrderf(expectedOrder int, actualOrder int) error {
	return fmt.Errorf(
		"%w: step order must be %d got %d",
		ErrValidation,
		expectedOrder,
		actualOrder,
	)
}
