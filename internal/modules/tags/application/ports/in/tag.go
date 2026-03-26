package in

import (
	"context"

	"recipes-desk/internal/modules/tags/application/dto"
	"recipes-desk/pkg/pagination"
)

type TagService interface {
	// SearchTags searches tags by name.
	SearchTags(
		ctx context.Context,
		query string,
		pagnreq pagination.Request,
	) (*pagination.Result[dto.Tag], error)

	// GetTagByID returns a tag by its ID.
	GetTagByID(ctx context.Context, id string) (*dto.Tag, error)

	// CreateTag creates a new tag if it doesn't exist.
	CreateTag(ctx context.Context, name string) (*dto.Tag, error)

	// EnsureTagExists ensures a tag exists, creating it if necessary.
	EnsureTagExists(ctx context.Context, tagName string) error

	// EnsureTagsExist ensures multiple tags exist.
	EnsureTagsExist(ctx context.Context, tagNames []string) error
}
