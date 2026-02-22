package handler

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/recipes/domain"
)

// toDomainRecipe converts request to domain model.
func toDomainRecipe(req CreateRequest) *domain.Recipe {
	ingredients := make([]domain.Ingredient, len(req.Ingredients))
	for i, ing := range req.Ingredients {
		ingredients[i] = domain.Ingredient{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}

	steps := make([]domain.Step, len(req.Steps))
	for i, step := range req.Steps {
		steps[i] = domain.Step{
			Order:       step.Order,
			Description: step.Description,
			Duration:    step.Duration,
		}
	}

	return &domain.Recipe{
		ID:          primitive.NilObjectID,
		Title:       req.Title,
		Description: req.Description,
		Ingredients: ingredients,
		Steps:       steps,
		CookingTime: req.CookingTime,
		Portions:    req.Portions,
		Tags:        req.Tags,
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	}
}

// toResponse converts domain model to response.
func toResponse(recipe *domain.Recipe) RecipeResponse {
	ingredients := make([]IngredientResponse, len(recipe.Ingredients))
	for i, ing := range recipe.Ingredients {
		ingredients[i] = IngredientResponse{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}

	steps := make([]StepResponse, len(recipe.Steps))
	for i, step := range recipe.Steps {
		steps[i] = StepResponse{
			Order:       step.Order,
			Description: step.Description,
			Duration:    step.Duration,
		}
	}

	return RecipeResponse{
		ID:          recipe.ID.Hex(),
		Title:       recipe.Title,
		Description: recipe.Description,
		Ingredients: ingredients,
		Steps:       steps,
		CookingTime: recipe.CookingTime,
		Portions:    recipe.Portions,
		Tags:        recipe.Tags,
		CreatedAt:   recipe.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   recipe.UpdatedAt.Format(time.RFC3339),
	}
}

// toListResponse converts domain recipes to list response.
func toListResponse(recipes []domain.Recipe) ListResponse {
	response := ListResponse{
		Recipes: make([]RecipeResponse, len(recipes)),
		Count:   len(recipes),
	}
	for i, recipe := range recipes {
		response.Recipes[i] = toResponse(&recipe)
	}
	return response
}
