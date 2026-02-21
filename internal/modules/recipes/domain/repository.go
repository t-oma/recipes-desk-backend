package domain

import (
	"context"
	"errors"
	"fmt"
)

// Error categories (sentinel errors).
var (
	ErrNotFound   = errors.New("recipe not found")
	ErrValidation = errors.New("validation error")
)

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

// Repository defines the interface for recipe data access.
type Repository interface {
	// Create inserts a new recipe
	Create(ctx context.Context, recipe *Recipe) error

	// FindByID finds a recipe by its ID
	FindByID(ctx context.Context, id string) (*Recipe, error)

	// FindAll returns all recipes
	FindAll(ctx context.Context) ([]Recipe, error)

	// Search searches recipes by title (case-insensitive)
	Search(ctx context.Context, query string) ([]Recipe, error)

	// Update updates an existing recipe
	Update(ctx context.Context, recipe *Recipe) error

	// Delete removes a recipe by its ID
	Delete(ctx context.Context, id string) error
}
