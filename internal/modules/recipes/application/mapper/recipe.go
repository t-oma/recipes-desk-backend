package mapper

import (
	"recipes-desk/internal/modules/recipes/application/dto"
	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

func ToRecipeDTO(recipe *entity.Recipe) *dto.Recipe {
	recipeIngredients := recipe.Ingredients()
	ingredients := make([]dto.Ingredient, len(recipeIngredients))
	for i, ing := range recipeIngredients {
		ingredients[i] = dto.Ingredient{
			Name:   ing.Name(),
			Amount: ing.Amount().Value(),
			Unit:   ing.Unit().Name(),
		}
	}

	recipeSteps := recipe.Steps()
	steps := make([]dto.Step, len(recipeSteps))
	for i, step := range recipeSteps {
		steps[i] = dto.Step{
			Order:       step.Order(),
			Description: step.Description(),
			Duration:    step.Duration(),
		}
	}

	recipeTags := recipe.Tags()
	tags := make([]string, len(recipeTags))
	for i, tag := range recipeTags {
		tags[i] = tag.String()
	}

	return &dto.Recipe{
		ID:          recipe.ID().String(),
		Title:       recipe.Title().String(),
		Description: recipe.Description().String(),
		Ingredients: ingredients,
		Steps:       steps,
		CookingTime: recipe.CookingTime().SecondsInt64(),
		Portions:    recipe.Portions().Value(),
		Tags:        tags,
		CreatedAt:   recipe.CreatedAt(),
		UpdatedAt:   recipe.UpdatedAt(),
	}
}

func ToDomainIngredient(ingredient dto.Ingredient) (valueobject.Ingredient, error) {
	ingr, err := valueobject.NewIngredient(ingredient.Name, ingredient.Amount, ingredient.Unit)
	if err != nil {
		return ingr, err
	}
	return ingr, nil
}

func ToDomainStep(step dto.Step) (valueobject.Step, error) {
	stepVO, err := valueobject.NewStep(step.Order, step.Description, int64(step.Duration.Seconds()))
	if err != nil {
		return stepVO, err
	}
	return stepVO, nil
}
