package httphandler

// CreateRecipeRequest represents the create/update request.
type CreateRecipeRequest struct {
	Title       string              `json:"title"       binding:"required"`
	Description string              `json:"description" binding:"required"`
	Ingredients []IngredientRequest `json:"ingredients" binding:"required,min=1"`
	Steps       []StepRequest       `json:"steps"       binding:"required,min=1"`
	CookingTime int64               `json:"cookingTime" binding:"required,min=1"`
	Portions    int                 `json:"portions"    binding:"required,min=1,max=100"`
	Tags        []string            `json:"tags"        binding:"required,min=1"`
}

// IngredientRequest represents an ingredient in the request.
type IngredientRequest struct {
	Name   string  `json:"name"   binding:"required"`
	Amount float64 `json:"amount" binding:"required,min=0"`
	Unit   string  `json:"unit"   binding:"required"`
}

// StepRequest represents a step in the request.
type StepRequest struct {
	Order       int    `json:"order"       binding:"required,min=1"`
	Description string `json:"description" binding:"required"`
	Duration    int64  `json:"duration"    binding:"min=0"`
}
