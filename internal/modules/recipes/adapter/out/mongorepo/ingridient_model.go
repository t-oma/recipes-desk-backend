package mongorepo

import (
	"recipes-desk/internal/modules/recipes/domain/recipe"
)

type ingredientModel struct {
	Name   string  `bson:"name"`
	Amount float64 `bson:"amount"`
	Unit   string  `bson:"unit"`
}

func (m *ingredientModel) toDomain() recipe.Ingredient {
	ingredient, _ := recipe.NewIngredient(m.Name, m.Amount, m.Unit)
	return ingredient
}

func ingredientModelFromDomain(ingredient recipe.Ingredient) ingredientModel {
	return ingredientModel{
		Name:   ingredient.Name(),
		Amount: ingredient.Amount().Value(),
		Unit:   ingredient.Unit().Name(),
	}
}
