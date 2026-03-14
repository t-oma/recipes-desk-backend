package httphandler

import (
	"recipes-desk/internal/modules/recipes/application/dto"
	"recipes-desk/pkg/pagination"
)

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

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

type PaginatedResponse[T any] struct {
	Items      []T            `json:"items"`
	Pagination PaginationMeta `json:"pagination"`
}

func toPaginatedResponse(result *pagination.Result[dto.Recipe]) PaginatedResponse[RecipeResponse] {
	items := make([]RecipeResponse, len(result.Items))
	for i, recipe := range result.Items {
		items[i] = *toRecipeResponse(&recipe)
	}

	return PaginatedResponse[RecipeResponse]{
		Items:      items,
		Pagination: PaginationMeta(result.Pagination),
	}
}
