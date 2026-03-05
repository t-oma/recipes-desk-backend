package mapper

import (
	"recipes-desk/internal/modules/recipes/application/dto"
	"recipes-desk/internal/modules/recipes/domain"
)

func ToRecipeDTO(recipe *domain.Recipe) *dto.Recipe {
	ingredients := make([]dto.Ingredient, len(recipe.Ingredients))
	for i, ing := range recipe.Ingredients {
		ingredients[i] = dto.Ingredient{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}

	steps := make([]dto.Step, len(recipe.Steps))
	for i, step := range recipe.Steps {
		steps[i] = dto.Step{
			Order:       step.Order,
			Description: step.Description,
			Duration:    step.Duration,
		}
	}

	return &dto.Recipe{
		ID:          recipe.ID,
		Title:       recipe.Title,
		Description: recipe.Description,
		Ingredients: ingredients,
		Steps:       steps,
		CookingTime: recipe.CookingTime,
		Portions:    recipe.Portions,
		Tags:        recipe.Tags,
		CreatedAt:   recipe.CreatedAt,
		UpdatedAt:   recipe.UpdatedAt,
	}
}
