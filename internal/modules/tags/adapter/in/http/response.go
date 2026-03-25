package httphandler

import (
	"time"

	"recipes-desk/internal/modules/tags/application/dto"
	"recipes-desk/pkg/pagination"
)

// TagResponse represents a tag in the response.
type TagResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// PaginationMeta contains pagination metadata for paginated responses.
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

func toTagResponse(tag *dto.Tag) TagResponse {
	return TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedAt: tag.CreatedAt.Format(time.RFC3339),
		UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
	}
}

func toPaginatedResponse(result *pagination.Result[dto.Tag]) PaginatedResponse[TagResponse] {
	items := make([]TagResponse, len(result.Items))
	for i, tag := range result.Items {
		items[i] = toTagResponse(&tag)
	}

	return PaginatedResponse[TagResponse]{
		Items: items,
		Pagination: PaginationMeta{
			Page:       result.Pagination.Page,
			Limit:      result.Pagination.Limit,
			Total:      result.Pagination.Total,
			TotalPages: result.Pagination.TotalPages,
			HasNext:    result.HasNext(),
			HasPrev:    result.HasPrev(),
		},
	}
}
