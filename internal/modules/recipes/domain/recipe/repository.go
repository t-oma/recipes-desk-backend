package recipe

import (
	"context"
	"errors"
)

// Repository defines the interface for recipe data access.
type Repository interface {
	// Create inserts a new recipe
	Create(ctx context.Context, recipe *Entity) (*Entity, error)

	// FindByID finds a recipe by its ID
	FindByID(ctx context.Context, id string) (*Entity, error)

	// FindAll returns all recipes
	FindAll(ctx context.Context) ([]Entity, error)

	// Search searches recipes by title (case-insensitive)
	Search(ctx context.Context, query string) ([]Entity, error)

	// Update updates an existing recipe
	Update(ctx context.Context, recipe *Entity) (*Entity, error)

	// Delete removes a recipe by its ID
	Delete(ctx context.Context, id string) error
}

var ErrNotFound = errors.New("recipe not found")
