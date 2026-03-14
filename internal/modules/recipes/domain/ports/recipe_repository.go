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

	// FindAll returns paginated recipes
	FindAll(ctx context.Context, skip, limit int64) ([]entity.Recipe, int64, error)

	// Search searches recipes by title (case-insensitive) with pagination
	Search(ctx context.Context, query string, skip, limit int64) ([]entity.Recipe, int64, error)

	// Update updates an existing recipe
	Update(ctx context.Context, recipe *entity.Recipe) (*entity.Recipe, error)

	// Delete removes a recipe by its ID
	Delete(ctx context.Context, id string) error
}
