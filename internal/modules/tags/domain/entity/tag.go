// Package entity contains tag entity definitions.
package entity

import (
	"time"

	vo "recipes-desk/internal/modules/tags/domain/valueobject"
)

// Tag represents a tag definition in the system.
type Tag struct {
	id        vo.TagID
	name      vo.TagName
	slug      vo.TagSlug
	createdAt time.Time
	updatedAt time.Time
}

// NewTag creates a new Tag.
func NewTag(id vo.TagID, name vo.TagName) (*Tag, error) {
	slug, err := vo.NewTagSlug(name.String())
	if err != nil {
		return nil, err
	}

	return &Tag{
		id:        id,
		name:      name,
		slug:      slug,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

// TagFrom creates a Tag from persistence data.
func TagFrom(
	id vo.TagID,
	name vo.TagName,
	slug vo.TagSlug,
	createdAt time.Time,
	updatedAt time.Time,
) *Tag {
	return &Tag{
		id:        id,
		name:      name,
		slug:      slug,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (t Tag) ID() vo.TagID {
	return t.id
}

func (t Tag) Name() vo.TagName {
	return t.name
}

func (t Tag) Slug() vo.TagSlug {
	return t.slug
}

func (t Tag) CreatedAt() time.Time {
	return t.createdAt
}

func (t Tag) UpdatedAt() time.Time {
	return t.updatedAt
}
