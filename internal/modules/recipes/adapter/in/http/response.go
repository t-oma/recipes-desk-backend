package httphandler

type RecipeResponse struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Ingredients []IngredientResponse `json:"ingredients"`
	Steps       []StepResponse       `json:"steps"`
	CookingTime int64                `json:"cookingTime"`
	Portions    int                  `json:"portions"`
	Tags        []string             `json:"tags"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
}

type IngredientResponse struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

type StepResponse struct {
	Order       int    `json:"order"`
	Description string `json:"description"`
	Duration    int64  `json:"duration"`
}
