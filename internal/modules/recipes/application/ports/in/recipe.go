package in

import (
	"context"

	"recipes-desk/internal/modules/recipes/application/dto"
	"recipes-desk/pkg/pagination"
)

// RecipeService defines the service interface.
type RecipeService interface {
	// Create creates a new recipe.
	Create(ctx context.Context, recipe dto.CreateRecipeInput) (*dto.Recipe, error)

	// GetByID returns a recipe by its ID.
	GetByID(ctx context.Context, id string) (*dto.Recipe, error)

	// GetAll returns all recipes with pagination.
	GetAll(ctx context.Context, pagnreq pagination.Request) (*pagination.Result[dto.Recipe], error)

	// Search searches recipes by title (case-insensitive) with pagination.
	Search(
		ctx context.Context,
		query string,
		pagnreq pagination.Request,
	) (*pagination.Result[dto.Recipe], error)

	// Update updates a recipe if the user has permission.
	Update(
		ctx context.Context,
		userID string,
		id string,
		recipe dto.UpdateRecipeInput,
	) (*dto.Recipe, error)

	// Delete removes a recipe by its ID if the user has permission.
	Delete(ctx context.Context, userID string, id string) error
}
