//go:build integration
// +build integration

package mongorepo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/tags/adapter/out/mongorepo"
	"recipes-desk/internal/modules/tags/domain"
	"recipes-desk/internal/modules/tags/domain/fixtures"
	"recipes-desk/pkg/testutils"
)

const _testDBName = "test_tags_integration"

func TestIntegration_TagRepository_Create(t *testing.T) {
	container, cleanup := testutils.SetupMongoDBContainer(t)
	defer cleanup()

	database := container.Client.Database(_testDBName)
	repo := mongorepo.NewTags(database)
	ctx := context.Background()

	require.NoError(t, repo.InitIndexes(ctx))

	t.Run("create new tag", func(t *testing.T) {
		tag := fixtures.NewTag(t, primitive.NewObjectID().Hex(), "Italian")
		created, err := repo.Create(ctx, tag)
		require.NoError(t, err)

		assert.Equal(t, tag.ID().String(), created.ID().String())
		assert.Equal(t, "Italian", created.Name().String())
		assert.Equal(t, "italian", created.Slug().String())
		assert.False(t, created.CreatedAt().IsZero())
		assert.False(t, created.UpdatedAt().IsZero())
	})

	t.Run("duplicate slug returns conflict", func(t *testing.T) {
		tag1 := fixtures.NewTag(t, primitive.NewObjectID().Hex(), "Spicy")
		_, err := repo.Create(ctx, tag1)
		require.NoError(t, err)

		tag2 := fixtures.NewTag(t, primitive.NewObjectID().Hex(), "spicy")
		_, err = repo.Create(ctx, tag2)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrConflict)
	})
}

func TestIntegration_TagRepository_FindByID(t *testing.T) {
	container, cleanup := testutils.SetupMongoDBContainer(t)
	defer cleanup()

	database := container.Client.Database(_testDBName)
	repo := mongorepo.NewTags(database)
	ctx := context.Background()

	require.NoError(t, repo.InitIndexes(ctx))

	tag := fixtures.NewTag(t, primitive.NewObjectID().Hex(), "Italian")
	created, err := repo.Create(ctx, tag)
	require.NoError(t, err)

	t.Run("find existing tag", func(t *testing.T) {
		found, err := repo.FindByID(ctx, created.ID().String())
		require.NoError(t, err)

		assert.Equal(t, created.ID().String(), found.ID().String())
		assert.Equal(t, "Italian", found.Name().String())
		assert.Equal(t, "italian", found.Slug().String())
	})

	t.Run("find non-existing tag", func(t *testing.T) {
		_, err := repo.FindByID(ctx, primitive.NewObjectID().Hex())
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("invalid object id returns not found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "invalid-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestIntegration_TagRepository_Search(t *testing.T) {
	container, cleanup := testutils.SetupMongoDBContainer(t)
	defer cleanup()

	database := container.Client.Database(_testDBName)
	repo := mongorepo.NewTags(database)
	ctx := context.Background()

	require.NoError(t, repo.InitIndexes(ctx))

	// Seed data
	tags := []struct {
		name string
	}{
		{"Italian"},
		{"Indian"},
		{"Thai"},
		{"Breakfast"},
		{"Dinner"},
	}

	for _, tt := range tags {
		tag := fixtures.NewTag(t, primitive.NewObjectID().Hex(), tt.name)
		_, err := repo.Create(ctx, tag)
		require.NoError(t, err)
	}

	t.Run("search by partial name", func(t *testing.T) {
		results, total, err := repo.Search(ctx, "ian", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
	})

	t.Run("search is case insensitive", func(t *testing.T) {
		results, total, err := repo.Search(ctx, "ITALIAN", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "Italian", results[0].Name().String())
	})

	t.Run("search empty query returns all", func(t *testing.T) {
		results, total, err := repo.Search(ctx, "", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, results, 5)
	})

	t.Run("search with pagination", func(t *testing.T) {
		results, total, err := repo.Search(ctx, "", 0, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, results, 2)

		results2, total2, err := repo.Search(ctx, "", 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total2)
		assert.Len(t, results2, 2)
	})

	t.Run("search no match", func(t *testing.T) {
		results, total, err := repo.Search(ctx, "zzzzz", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, results, 0)
	})
}
