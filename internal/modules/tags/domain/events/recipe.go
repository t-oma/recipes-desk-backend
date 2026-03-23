package events

type RecipeCreated struct {
	RecipeID string   `json:"recipeId"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
}

type RecipeDeleted struct {
	RecipeID string   `json:"recipeId"`
	Tags     []string `json:"tags"`
}
