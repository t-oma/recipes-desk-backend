//go:build integration
// +build integration

package mongorepo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/recipes/adapter/out/mongorepo"
	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/fixtures"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
	"recipes-desk/pkg/testutils"
)

const _testDBName = "test_recipes_integration"

func TestIntegration_RecipeRepository_Create(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	t.Run("create new recipe", func(t *testing.T) {
		entity := fixtures.NewRecipe(
			t,
			primitive.NewObjectID().Hex(),
			primitive.NewObjectID().Hex(),
			"Integration Test Recipe",
		)
		created, err := repo.Create(ctx, entity)
		require.NoError(t, err)

		// Verify ID preserved
		assert.Equal(t, entity.ID().String(), created.ID().String())
		// Verify timestamps were set
		assert.False(t, created.CreatedAt().IsZero())
		assert.False(t, created.UpdatedAt().IsZero())
	})
}

func TestIntegration_RecipeRepository_FindByID(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create a recipe first
	entity := fixtures.NewRecipe(
		t,
		primitive.NewObjectID().Hex(),
		primitive.NewObjectID().Hex(),
		"Test Recipe",
	)
	created, err := repo.Create(ctx, entity)
	require.NoError(t, err)
	id := created.ID().String()

	t.Run("find existing recipe", func(t *testing.T) {
		found, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		require.NotEmpty(t, found.ID().String())
		assert.Equal(t, id, found.ID().String())
		assert.Equal(t, entity.Title().String(), found.Title().String())
	})

	t.Run("find non-existing recipe", func(t *testing.T) {
		nonExistingID := primitive.NewObjectID().Hex()
		_, err := repo.FindByID(ctx, nonExistingID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("find with invalid id", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "invalid-id")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestIntegration_RecipeRepository_FindAll(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	const totalRecipes = 3
	for i := 0; i < totalRecipes; i++ {
		entity := fixtures.NewRecipe(
			t,
			primitive.NewObjectID().Hex(),
			primitive.NewObjectID().Hex(),
			fmt.Sprintf("Recipe %d", i),
		)
		entity, err := repo.Create(ctx, entity)
		require.NoError(t, err)
	}

	tests := []struct {
		name       string
		skip       int64
		limit      int64
		wantRecipe int
	}{
		{
			name:       "no skip",
			skip:       0,
			limit:      totalRecipes,
			wantRecipe: totalRecipes,
		},
		{
			name:       "skip 2 recipes",
			skip:       2,
			limit:      totalRecipes,
			wantRecipe: totalRecipes - 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipes, total, err := repo.FindAll(ctx, tt.skip, tt.limit)
			require.NoError(t, err)
			assert.Len(t, recipes, tt.wantRecipe)
			assert.Equal(t, int64(totalRecipes), total)
		})
	}
}

func TestIntegration_RecipeRepository_Search(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create recipes with different titles
	testRecipes := []struct {
		title string
		id    string
	}{
		{"Pasta Carbonara", ""},
		{"Pasta Bolognese", ""},
		{"Chicken Curry", ""},
	}

	for i, tr := range testRecipes {
		id, err := valueobject.NewRecipeID(primitive.NewObjectID().Hex())
		require.NoError(t, err)
		title, err := valueobject.NewTitle(tr.title)
		require.NoError(t, err)
		authorID, err := valueobject.NewAuthorID(primitive.NewObjectID().Hex())
		require.NoError(t, err)

		customEntity, err := entity.NewRecipe(
			id,
			title,
			fixtures.ValidDescription(t),
			fixtures.ValidIngredients(t),
			fixtures.ValidSteps(t),
			fixtures.ValidCookingTime(t),
			fixtures.ValidPortions(t),
			fixtures.ValidTags(t),
			authorID,
		)
		require.NoError(t, err)

		created, err := repo.Create(ctx, customEntity)
		require.NoError(t, err)
		testRecipes[i].id = created.ID().String()
	}

	t.Run("search by title - case insensitive", func(t *testing.T) {
		results, _, err := repo.Search(ctx, "pasta", 0, 10)
		require.NoError(t, err)
		for _, result := range results {
			require.NotEmpty(t, result.ID().String())
		}
		assert.Len(t, results, 2)
	})

	t.Run("search with no matches", func(t *testing.T) {
		results, _, err := repo.Search(ctx, "sushi", 0, 10)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

func TestIntegration_RecipeRepository_Update(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create a recipe
	recipe := fixtures.NewRecipe(
		t,
		primitive.NewObjectID().Hex(),
		primitive.NewObjectID().Hex(),
		"Original Title",
	)
	created, err := repo.Create(ctx, recipe)
	require.NoError(t, err)
	id := created.ID()

	originalCreatedAt := created.CreatedAt()

	t.Run("update existing recipe", func(t *testing.T) {
		// Wait a bit to ensure UpdatedAt changes
		time.Sleep(10 * time.Millisecond)

		// Fetch the recipe first
		recipeToUpdate, err := repo.FindByID(ctx, id.String())
		require.NoError(t, err)

		// Create updated version with new title
		ingredients := recipeToUpdate.Ingredients()
		steps := recipeToUpdate.Steps()
		tags := recipeToUpdate.Tags()

		updatedTitle, err := valueobject.NewTitle("Updated Title")
		require.NoError(t, err)

		updatedEntity, err := entity.NewRecipe(
			id,
			updatedTitle,
			recipeToUpdate.Description(),
			ingredients,
			steps,
			recipeToUpdate.CookingTime(),
			recipeToUpdate.Portions(),
			tags,
			recipeToUpdate.AuthorID(),
		)
		require.NoError(t, err)
		updatedEntity.RestoreFromPersistence(
			recipeToUpdate.UpdatedAt(),
			recipeToUpdate.CreatedAt(),
		)

		result, err := repo.Update(ctx, updatedEntity)
		require.NoError(t, err)
		require.NotEmpty(t, result.ID().String())

		// Verify update
		updated, err := repo.FindByID(ctx, id.String())
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", updated.Title().String())
		// Compare timestamps in UTC to avoid timezone issues
		assert.WithinDuration(
			t,
			originalCreatedAt.UTC(),
			updated.CreatedAt().UTC(),
			time.Millisecond,
		) // CreatedAt preserved
		assert.True(t, updated.UpdatedAt().After(originalCreatedAt))
	})

	t.Run("update non-existing recipe", func(t *testing.T) {
		id, err := valueobject.NewRecipeID(primitive.NewObjectID().Hex())
		require.NoError(t, err)
		authorID, err := valueobject.NewAuthorID(primitive.NewObjectID().Hex())
		require.NoError(t, err)

		nonExisting, err := entity.NewRecipe(
			id,
			fixtures.ValidTitle(t),
			fixtures.ValidDescription(t),
			fixtures.ValidIngredients(t),
			fixtures.ValidSteps(t),
			fixtures.ValidCookingTime(t),
			fixtures.ValidPortions(t),
			fixtures.ValidTags(t),
			authorID,
		)
		require.NoError(t, err)

		_, err = repo.Update(ctx, nonExisting)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestIntegration_RecipeRepository_Delete(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create a recipe
	entity := fixtures.NewRecipe(
		t,
		primitive.NewObjectID().Hex(),
		primitive.NewObjectID().Hex(),
		"To be deleted",
	)
	created, err := repo.Create(ctx, entity)
	require.NoError(t, err)
	id := created.ID().String()

	t.Run("delete existing recipe", func(t *testing.T) {
		err := repo.Delete(ctx, id)
		require.NoError(t, err)

		// Verify it's gone
		_, err = repo.FindByID(ctx, id)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("delete non-existing recipe", func(t *testing.T) {
		nonExistingID := primitive.NewObjectID().Hex()
		err := repo.Delete(ctx, nonExistingID)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("delete with invalid id", func(t *testing.T) {
		err := repo.Delete(ctx, "invalid-id")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}
