package service

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

type RecipeDTO struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Ingredients []IngredientDTO `json:"ingredients"`
	Steps       []StepDTO       `json:"steps"`
	CookingTime int             `json:"cookingTime"`
	Portions    int             `json:"portions"`
	Tags        []string        `json:"tags"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
}

type IngredientDTO struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

type StepDTO struct {
	Order       int    `json:"order"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
}
