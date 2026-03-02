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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/repository/mongorepo"
)

func setupMongoContainer(t *testing.T) (*mongo.Database, func()) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx, "mongo:8",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("27017/tcp")),
	)
	require.NoError(t, err)

	// Get connection string
	connStr, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	require.NoError(t, err)

	// Check connection
	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	db := client.Database("test_recipes")

	cleanup := func() {
		client.Disconnect(ctx)
		mongoContainer.Terminate(ctx)
	}

	return db, cleanup
}

func TestIntegration_MongoRepository_Create(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	t.Run("create new recipe", func(t *testing.T) {
		recipe := &domain.Recipe{
			Title:       "Integration Test Recipe",
			Description: "This is a valid description for integration test",
			Ingredients: []domain.Ingredient{
				{Name: "Flour", Amount: 500, Unit: "g"},
				{Name: "Eggs", Amount: 3, Unit: "pcs"},
			},
			Steps: []domain.Step{
				{Order: 1, Description: "Mix ingredients", Duration: 5},
				{Order: 2, Description: "Bake", Duration: 30},
			},
			CookingTime: 35,
			Portions:    4,
			Tags:        []string{"test", "integration"},
		}

		recipe, err := repo.Create(ctx, recipe)
		require.NoError(t, err)

		// Verify ID was set
		assert.NotZero(t, recipe.ID)
		// Verify timestamps were set
		assert.NotZero(t, recipe.CreatedAt)
		assert.NotZero(t, recipe.UpdatedAt)
	})
}

func TestIntegration_MongoRepository_FindByID(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create a recipe first
	recipe := &domain.Recipe{
		Title:       "Find Me",
		Description: "This recipe will be found by ID",
		Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
		Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
		CookingTime: 10,
		Portions:    2,
		Tags:        []string{"test"},
	}
	recipe, err := repo.Create(ctx, recipe)
	require.NoError(t, err)

	t.Run("find existing recipe", func(t *testing.T) {
		found, err := repo.FindByID(ctx, recipe.ID)
		require.NoError(t, err)
		require.NotZero(t, found.ID)
		assert.Equal(t, recipe.ID, found.ID)
		assert.Equal(t, recipe.Title, found.Title)
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

func TestIntegration_MongoRepository_FindAll(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create multiple recipes
	for i := 0; i < 3; i++ {
		recipe := &domain.Recipe{
			Title:       fmt.Sprintf("Recipe %d", i),
			Description: "Test description",
			Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
			Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
			CookingTime: 10,
			Portions:    2,
			Tags:        []string{"test"},
		}
		recipe, err := repo.Create(ctx, recipe)
		require.NoError(t, err)
	}

	t.Run("find all recipes", func(t *testing.T) {
		recipes, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.Len(t, recipes, 3)
	})
}

func TestIntegration_MongoRepository_Search(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create recipes with different titles
	recipes := []*domain.Recipe{
		{
			Title:       "Pasta Carbonara",
			Description: "Classic Italian dish",
			Ingredients: []domain.Ingredient{{Name: "Pasta", Amount: 400, Unit: "g"}},
			Steps:       []domain.Step{{Order: 1, Description: "Cook", Duration: 10}},
			CookingTime: 20,
			Portions:    4,
			Tags:        []string{"italian"},
		},
		{
			Title:       "Pasta Bolognese",
			Description: "Italian meat sauce",
			Ingredients: []domain.Ingredient{{Name: "Pasta", Amount: 400, Unit: "g"}},
			Steps:       []domain.Step{{Order: 1, Description: "Cook", Duration: 10}},
			CookingTime: 30,
			Portions:    4,
			Tags:        []string{"italian"},
		},
		{
			Title:       "Chicken Curry",
			Description: "Spicy Indian dish",
			Ingredients: []domain.Ingredient{{Name: "Chicken", Amount: 500, Unit: "g"}},
			Steps:       []domain.Step{{Order: 1, Description: "Cook", Duration: 20}},
			CookingTime: 45,
			Portions:    4,
			Tags:        []string{"indian", "spicy"},
		},
	}

	for i, recipe := range recipes {
		var err error
		recipes[i], err = repo.Create(ctx, recipe)
		require.NoError(t, err)
	}

	t.Run("search by title - case insensitive", func(t *testing.T) {
		results, err := repo.Search(ctx, "pasta")
		require.NoError(t, err)
		for _, result := range results {
			require.NotZero(t, result.ID)
		}
		assert.Len(t, results, 2)
	})

	t.Run("search with no matches", func(t *testing.T) {
		results, err := repo.Search(ctx, "sushi")
		require.NoError(t, err)
		for _, result := range results {
			require.NotZero(t, result.ID)
		}
		assert.Len(t, results, 0)
	})
}

func TestIntegration_MongoRepository_Update(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create a recipe
	recipe := &domain.Recipe{
		Title:       "Original Title",
		Description: "Original description",
		Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
		Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
		CookingTime: 10,
		Portions:    2,
		Tags:        []string{"test"},
	}
	recipe, err := repo.Create(ctx, recipe)
	require.NoError(t, err)

	originalCreatedAt := recipe.CreatedAt

	t.Run("update existing recipe", func(t *testing.T) {
		// Wait a bit to ensure UpdatedAt changes
		time.Sleep(10 * time.Millisecond)

		recipe.Title = "Updated Title"
		recipe, err := repo.Update(ctx, recipe)
		require.NoError(t, err)
		require.NotZero(t, recipe.ID)

		// Verify update
		updated, err := repo.FindByID(ctx, recipe.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", updated.Title)
		// Compare timestamps in UTC to avoid timezone issues
		assert.WithinDuration(
			t,
			originalCreatedAt.UTC(),
			updated.CreatedAt.UTC(),
			time.Millisecond,
		) // CreatedAt preserved
		assert.True(t, updated.UpdatedAt.After(originalCreatedAt))
	})

	t.Run("update non-existing recipe", func(t *testing.T) {
		nonExisting := &domain.Recipe{
			ID:          primitive.NewObjectID().Hex(),
			Title:       "Does not exist",
			Description: "This recipe doesn't exist",
			Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
			Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
			CookingTime: 10,
			Portions:    2,
			Tags:        []string{"test"},
		}

		_, err := repo.Update(ctx, nonExisting)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestIntegration_MongoRepository_Delete(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRecipes(db)
	ctx := context.Background()

	// Create a recipe
	recipe := &domain.Recipe{
		Title:       "To be deleted",
		Description: "This recipe will be deleted",
		Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
		Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
		CookingTime: 10,
		Portions:    2,
		Tags:        []string{"test"},
	}
	recipe, err := repo.Create(ctx, recipe)
	require.NoError(t, err)

	t.Run("delete existing recipe", func(t *testing.T) {
		err := repo.Delete(ctx, recipe.ID)
		require.NoError(t, err)

		// Verify it's gone
		_, err = repo.FindByID(ctx, recipe.ID)
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
