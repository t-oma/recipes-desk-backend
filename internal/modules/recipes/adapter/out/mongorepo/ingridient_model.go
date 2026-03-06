package mongorepo

import "recipes-desk/internal/modules/recipes/domain/valueobject"

type ingredientModel struct {
	Name   string  `bson:"name"`
	Amount float64 `bson:"amount"`
	Unit   string  `bson:"unit"`
}

func (m *ingredientModel) toDomain() (valueobject.Ingredient, error) {
	ingredient, err := valueobject.NewIngredient(m.Name, m.Amount, m.Unit)
	return ingredient, err
}

func ingredientModelFromDomain(ingredient valueobject.Ingredient) ingredientModel {
	return ingredientModel{
		Name:   ingredient.Name(),
		Amount: ingredient.Amount().Value(),
		Unit:   ingredient.Unit().Name(),
	}
}
