// Package ports defines repository interfaces for tags.
package ports

import (
	"context"

	"recipes-desk/internal/modules/tags/domain/entity"
)

// TagRepository defines the interface for tag data access.
type TagRepository interface {
	// Create inserts a new tag.
	Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error)

	// FindByID finds a tag by its ID.
	FindByID(ctx context.Context, id string) (*entity.Tag, error)

	// Search finds tags by name (case-insensitive, partial match).
	Search(ctx context.Context, query string, skip, limit int64) ([]entity.Tag, int64, error)
}
