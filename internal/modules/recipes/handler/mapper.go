package handler

import (
	"recipes-desk/internal/modules/recipes/service"
)

func mapIngredients(ingredients []IngredientRequest) []service.IngredientInput {
	mapped := make([]service.IngredientInput, len(ingredients))
	for i, ing := range ingredients {
		mapped[i] = service.IngredientInput{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	return mapped
}

func mapSteps(steps []StepRequest) []service.StepInput {
	mapped := make([]service.StepInput, len(steps))
	for i, step := range steps {
		mapped[i] = service.StepInput{
			Order:       step.Order,
			Description: step.Description,
			Duration:    step.Duration,
		}
	}
	return mapped
}
