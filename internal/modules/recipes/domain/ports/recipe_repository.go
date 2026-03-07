package ports

import (
	"context"

	"recipes-desk/internal/modules/recipes/domain/entity"
)

// RecipeRepository defines the interface for recipe data access.
type RecipeRepository interface {
	// Create inserts a new recipe
	Create(ctx context.Context, recipe *entity.Recipe) (*entity.Recipe, error)

	// FindByID finds a recipe by its ID
	FindByID(ctx context.Context, id string) (*entity.Recipe, error)

	// FindAll returns all recipes
	FindAll(ctx context.Context) ([]entity.Recipe, error)

	// Search searches recipes by title (case-insensitive)
	Search(ctx context.Context, query string) ([]entity.Recipe, error)

	// Update updates an existing recipe
	Update(ctx context.Context, recipe *entity.Recipe) (*entity.Recipe, error)

	// Delete removes a recipe by its ID
	Delete(ctx context.Context, id string) error
}
