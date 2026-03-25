// Package fixtures provides helpers for creating tag domain objects in tests.
package fixtures

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/tags/domain/entity"
	vo "recipes-desk/internal/modules/tags/domain/valueobject"
)

// MustTagID creates a TagID or fails the test.
func MustTagID(t *testing.T, id string) vo.TagID {
	t.Helper()

	v, err := vo.NewTagID(id)
	require.NoError(t, err)
	return v
}

// MustTagName creates a TagName or fails the test.
func MustTagName(t *testing.T, name string) vo.TagName {
	t.Helper()

	v, err := vo.NewTagName(name)
	require.NoError(t, err)
	return v
}

// MustTagSlug creates a TagSlug or fails the test.
func MustTagSlug(t *testing.T, slug string) vo.TagSlug {
	t.Helper()

	v, err := vo.NewTagSlug(slug)
	require.NoError(t, err)
	return v
}

// NewTag creates a valid Tag entity for testing.
func NewTag(t *testing.T, id, name string) *entity.Tag {
	t.Helper()

	tag, err := entity.NewTag(
		MustTagID(t, id),
		MustTagName(t, name),
	)
	require.NoError(t, err)
	return tag
}

// NewTagFrom creates a Tag from persistence data for testing.
func NewTagFrom(t *testing.T, id, name, slug string, createdAt, updatedAt time.Time) *entity.Tag {
	t.Helper()

	return entity.TagFrom(
		MustTagID(t, id),
		MustTagName(t, name),
		MustTagSlug(t, slug),
		createdAt,
		updatedAt,
	)
}
