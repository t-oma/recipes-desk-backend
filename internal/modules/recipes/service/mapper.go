package service

import (
	"time"

	"recipes-desk/internal/modules/recipes/domain"
)

func toRecipeDTO(recipe *domain.Recipe) *RecipeDTO {
	ingredients := make([]IngredientDTO, len(recipe.Ingredients))
	for i, ing := range recipe.Ingredients {
		ingredients[i] = IngredientDTO{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}

	steps := make([]StepDTO, len(recipe.Steps))
	for i, step := range recipe.Steps {
		steps[i] = StepDTO{
			Order:       step.Order,
			Description: step.Description,
			Duration:    step.Duration,
		}
	}

	return &RecipeDTO{
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
