package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/tags/domain/entity"
	vo "recipes-desk/internal/modules/tags/domain/valueobject"
)

func TestNewTag(t *testing.T) {
	tests := []struct {
		name string
		id   string
		tag  string
	}{
		{
			name: "creates tag with auto-generated slug",
			id:   "abc123",
			tag:  "Italian",
		},
		{
			name: "slug is lowercased",
			id:   "def456",
			tag:  "Gluten-Free",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := vo.NewTagID(tt.id)
			require.NoError(t, err)
			name, err := vo.NewTagName(tt.tag)
			require.NoError(t, err)

			tag, err := entity.NewTag(id, name)
			require.NoError(t, err)

			assert.Equal(t, tt.id, tag.ID().String())
			assert.Equal(t, tt.tag, tag.Name().String())
			assert.False(t, tag.CreatedAt().IsZero())
			assert.False(t, tag.UpdatedAt().IsZero())
		})
	}
}

func TestTagFrom(t *testing.T) {
	id, _ := vo.NewTagID("id1")
	name, _ := vo.NewTagName("Italian")
	slug, _ := vo.NewTagSlug("italian")
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	tag := entity.TagFrom(id, name, slug, createdAt, updatedAt)

	assert.Equal(t, "id1", tag.ID().String())
	assert.Equal(t, "Italian", tag.Name().String())
	assert.Equal(t, "italian", tag.Slug().String())
	assert.Equal(t, createdAt, tag.CreatedAt())
	assert.Equal(t, updatedAt, tag.UpdatedAt())
}
