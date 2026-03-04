package httphandler

import (
	"time"

	"recipes-desk/internal/modules/recipes/application/dto"
)

func mapIngredients(ingredients []IngredientRequest) []dto.IngredientInput {
	mapped := make([]dto.IngredientInput, len(ingredients))
	for i, ing := range ingredients {
		mapped[i] = dto.IngredientInput{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	return mapped
}

func mapSteps(steps []StepRequest) []dto.StepInput {
	mapped := make([]dto.StepInput, len(steps))
	for i, step := range steps {
		mapped[i] = dto.StepInput{
			Order:       step.Order,
			Description: step.Description,
			Duration:    step.Duration,
		}
	}
	return mapped
}

func toRecipeResponse(recipe *dto.Recipe) *RecipeResponse {
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

	return &RecipeResponse{
		ID:          recipe.ID,
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
