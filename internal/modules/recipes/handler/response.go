package handler

// RecipeResponse represents a recipe response.
type RecipeResponse struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Ingredients []IngredientResponse `json:"ingredients"`
	Steps       []StepResponse       `json:"steps"`
	CookingTime int                  `json:"cookingTime"`
	Portions    int                  `json:"portions"`
	Tags        []string             `json:"tags"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
}

// IngredientResponse represents an ingredient in the response.
type IngredientResponse struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

// StepResponse represents a step in the response.
type StepResponse struct {
	Order       int    `json:"order"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
}

// ListResponse represents the list response.
type ListResponse struct {
	Recipes []RecipeResponse `json:"recipes"`
	Count   int              `json:"count"`
}
