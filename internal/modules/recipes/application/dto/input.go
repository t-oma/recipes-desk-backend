package dto

type CreateRecipeInput struct {
	Title       string
	Description string
	Ingredients []IngredientInput
	Steps       []StepInput
	CookingTime int
	Portions    int
	Tags        []string
}

type IngredientInput struct {
	Name   string
	Amount float64
	Unit   string
}

type StepInput struct {
	Order       int
	Description string
	Duration    int
}

type UpdateRecipeInput struct {
	Title       string
	Description string
	Ingredients []IngredientInput
	Steps       []StepInput
	CookingTime int
	Portions    int
	Tags        []string
}
