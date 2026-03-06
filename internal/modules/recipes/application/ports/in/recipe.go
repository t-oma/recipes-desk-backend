package in

import (
	"context"

	"recipes-desk/internal/modules/recipes/application/dto"
)

// RecipeService defines the service interface.
type RecipeService interface {
	Create(ctx context.Context, recipe dto.CreateRecipeInput) (*dto.Recipe, error)
	GetByID(ctx context.Context, id string) (*dto.Recipe, error)
	GetAll(ctx context.Context) ([]dto.Recipe, error)
	Search(ctx context.Context, query string) ([]dto.Recipe, error)
	Update(
		ctx context.Context,
		userID string,
		id string,
		recipe dto.UpdateRecipeInput,
	) (*dto.Recipe, error)
	Delete(ctx context.Context, id string) error
}
