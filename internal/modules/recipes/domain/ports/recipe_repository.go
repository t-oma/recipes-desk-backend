package ports

import (
	"context"
	"errors"

	"recipes-desk/internal/modules/recipes/domain/recipe"
)

// RecipeRepository defines the interface for recipe data access.
type RecipeRepository interface {
	// Create inserts a new recipe
	Create(ctx context.Context, recipe *recipe.Entity) (*recipe.Entity, error)

	// FindByID finds a recipe by its ID
	FindByID(ctx context.Context, id string) (*recipe.Entity, error)

	// FindAll returns all recipes
	FindAll(ctx context.Context) ([]recipe.Entity, error)

	// Search searches recipes by title (case-insensitive)
	Search(ctx context.Context, query string) ([]recipe.Entity, error)

	// Update updates an existing recipe
	Update(ctx context.Context, recipe *recipe.Entity) (*recipe.Entity, error)

	// Delete removes a recipe by its ID
	Delete(ctx context.Context, id string) error
}

var ErrNotFound = errors.New("recipe not found")
