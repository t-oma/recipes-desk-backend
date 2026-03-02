package mongorepo

import "recipes-desk/internal/modules/recipes/domain"

type ingredientModel struct {
	Name   string  `bson:"name"`
	Amount float64 `bson:"amount"`
	Unit   string  `bson:"unit"`
}

func (m *ingredientModel) toDomain() domain.Ingredient {
	return domain.Ingredient{
		Name:   m.Name,
		Amount: m.Amount,
		Unit:   m.Unit,
	}
}

func ingredientModelFromDomain(ingredient domain.Ingredient) ingredientModel {
	return ingredientModel{
		Name:   ingredient.Name,
		Amount: ingredient.Amount,
		Unit:   ingredient.Unit,
	}
}
