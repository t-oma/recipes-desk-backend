package events

// RecipeCreated represents a recipe.created domain event.
type RecipeCreated struct {
	RecipeID string   `json:"recipeId"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
}

// RecipeDeleted represents a recipe.deleted domain event.
type RecipeDeleted struct {
	RecipeID string   `json:"recipeId"`
	Tags     []string `json:"tags"`
}
