package domain

import (
	"context"
	"errors"
)

// RecipesRepository defines the interface for recipe data access.
type RecipesRepository interface {
	// Create inserts a new recipe
	Create(ctx context.Context, recipe *Recipe) (*Recipe, error)

	// FindByID finds a recipe by its ID
	FindByID(ctx context.Context, id string) (*Recipe, error)

	// FindAll returns all recipes
	FindAll(ctx context.Context) ([]Recipe, error)

	// Search searches recipes by title (case-insensitive)
	Search(ctx context.Context, query string) ([]Recipe, error)

	// Update updates an existing recipe
	Update(ctx context.Context, recipe *Recipe) (*Recipe, error)

	// Delete removes a recipe by its ID
	Delete(ctx context.Context, id string) error
}

// Error categories (sentinel errors).
var (
	ErrNotFound = errors.New("recipe not found")
)
