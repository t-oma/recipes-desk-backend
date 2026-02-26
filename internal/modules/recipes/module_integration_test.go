//go:build integration
// +build integration

package recipes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"recipes-desk/internal/modules/recipes"
	"recipes-desk/internal/modules/recipes/handler"
)

func setupTestContainer(t *testing.T) (*mongo.Database, func()) {
	ctx := context.Background()

	mongoContainer, err := mongodb.Run(ctx, "mongo:8",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("27017/tcp")),
	)
	require.NoError(t, err)

	connStr, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	db := client.Database("test_recipes_integration")

	cleanup := func() {
		client.Disconnect(ctx)
		mongoContainer.Terminate(ctx)
	}

	return db, cleanup
}

func setupModuleTest(t *testing.T) (*gin.Engine, *recipes.Module, func()) {
	db, cleanup := setupTestContainer(t)

	logger := zerolog.New(nil)
	module := recipes.NewModule(db, &logger)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	api := router.Group("/api/v1")

	public := api.Group("")
	protected := api.Group("")

	protected.Use(func(c *gin.Context) {
		c.Set("userID", "user-id")
		c.Next()
	})

	module.RegisterRoutes(public, protected)

	return router, module, cleanup
}

func TestIntegration_Module_FullCRUD(t *testing.T) {
	router, _, cleanup := setupModuleTest(t)
	defer cleanup()

	var createdRecipe handler.RecipeResponse
	recipeID := ""

	t.Run("create recipe", func(t *testing.T) {
		body := map[string]any{
			"title":       "Integration Test Recipe",
			"description": "This is a comprehensive integration test for the recipes module",
			"ingredients": []map[string]any{
				{"name": "Flour", "amount": 500, "unit": "g"},
				{"name": "Eggs", "amount": 3, "unit": "pcs"},
				{"name": "Milk", "amount": 250, "unit": "ml"},
			},
			"steps": []map[string]any{
				{"order": 1, "description": "Mix dry ingredients", "duration": 5},
				{"order": 2, "description": "Add wet ingredients", "duration": 3},
				{"order": 3, "description": "Bake in oven", "duration": 30},
			},
			"cookingTime": 45,
			"portions":    6,
			"tags":        []string{"integration", "test", "golang"},
		}

		bodyJSON, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/recipes", bytes.NewBuffer(bodyJSON))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		err := json.Unmarshal(w.Body.Bytes(), &createdRecipe)
		require.NoError(t, err)

		recipeID = createdRecipe.ID
		assert.NotEmpty(t, recipeID)
		assert.Equal(t, "Integration Test Recipe", createdRecipe.Title)
		assert.Equal(t, 45, createdRecipe.CookingTime)
		assert.Equal(t, 6, createdRecipe.Portions)
		assert.Len(t, createdRecipe.Ingredients, 3)
		assert.Len(t, createdRecipe.Steps, 3)
		assert.Len(t, createdRecipe.Tags, 3)
	})

	t.Run("get recipe by id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes/"+recipeID, nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var fetchedRecipe handler.RecipeResponse
		err := json.Unmarshal(w.Body.Bytes(), &fetchedRecipe)
		require.NoError(t, err)

		assert.Equal(t, recipeID, fetchedRecipe.ID)
		assert.Equal(t, createdRecipe.Title, fetchedRecipe.Title)
		assert.Equal(t, createdRecipe.Description, fetchedRecipe.Description)
	})

	t.Run("update recipe", func(t *testing.T) {
		// Wait a bit to ensure UpdatedAt changes
		time.Sleep(10 * time.Millisecond)

		updateBody := map[string]any{
			"title":       "Updated Integration Recipe",
			"description": "This recipe has been updated during integration testing",
			"ingredients": []map[string]any{
				{"name": "Flour", "amount": 600, "unit": "g"},
				{"name": "Eggs", "amount": 4, "unit": "pcs"},
			},
			"steps": []map[string]any{
				{"order": 1, "description": "Mix all ingredients", "duration": 10},
				{"order": 2, "description": "Bake", "duration": 35},
			},
			"cookingTime": 50,
			"portions":    8,
			"tags":        []string{"updated", "integration"},
		}

		bodyJSON, _ := json.Marshal(updateBody)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/recipes/"+recipeID, bytes.NewBuffer(bodyJSON))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedRecipe handler.RecipeResponse
		err := json.Unmarshal(w.Body.Bytes(), &updatedRecipe)
		require.NoError(t, err)

		assert.Equal(t, recipeID, updatedRecipe.ID)
		assert.Equal(t, "Updated Integration Recipe", updatedRecipe.Title)
		assert.Equal(t, 50, updatedRecipe.CookingTime)
		assert.Equal(t, 8, updatedRecipe.Portions)
	})

	t.Run("list all recipes and verify update", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var listResponse handler.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &listResponse)
		require.NoError(t, err)

		assert.Equal(t, 1, listResponse.Count)
		assert.Len(t, listResponse.Recipes, 1)
		assert.Equal(t, "Updated Integration Recipe", listResponse.Recipes[0].Title)
	})

	t.Run("delete recipe", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/recipes/"+recipeID, nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "recipe deleted successfully", response["message"])
	})

	t.Run("verify recipe is deleted", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes/"+recipeID, nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestIntegration_Module_Search(t *testing.T) {
	router, _, cleanup := setupModuleTest(t)
	defer cleanup()

	// Create test recipes
	recipes := []map[string]any{
		{
			"title":       "Pasta Carbonara",
			"description": "Classic Italian pasta dish with eggs and cheese",
			"ingredients": []map[string]any{{"name": "Pasta", "amount": 400, "unit": "g"}},
			"steps": []map[string]any{
				{"order": 1, "description": "Cook", "duration": 10},
			},
			"cookingTime": 20,
			"portions":    4,
			"tags":        []string{"italian"},
		},
		{
			"title":       "Pasta Bolognese",
			"description": "Italian pasta with meat sauce",
			"ingredients": []map[string]any{{"name": "Pasta", "amount": 400, "unit": "g"}},
			"steps": []map[string]any{
				{"order": 1, "description": "Cook", "duration": 15},
			},
			"cookingTime": 30,
			"portions":    4,
			"tags":        []string{"italian"},
		},
		{
			"title":       "Chicken Curry",
			"description": "Spicy Indian curry dish",
			"ingredients": []map[string]any{
				{"name": "Chicken", "amount": 500, "unit": "g"},
			},
			"steps": []map[string]any{
				{"order": 1, "description": "Cook", "duration": 20},
			},
			"cookingTime": 45,
			"portions":    4,
			"tags":        []string{"indian", "spicy"},
		},
	}

	// Create all recipes
	for _, recipe := range recipes {
		bodyJSON, _ := json.Marshal(recipe)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/recipes", bytes.NewBuffer(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	t.Run("search by title - case insensitive", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes/search?q=pasta", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handler.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Count)
		assert.Len(t, response.Recipes, 2)
	})

	t.Run("search with uppercase query", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes/search?q=PASTA", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handler.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Count)
	})

	t.Run("search with no matches", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes/search?q=sushi", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handler.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 0, response.Count)
		assert.Len(t, response.Recipes, 0)
	})

	t.Run("search partial match", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes/search?q=carbon", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handler.ListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, 1, response.Count)
		assert.Equal(t, "Pasta Carbonara", response.Recipes[0].Title)
	})
}

func TestIntegration_Module_ErrorCases(t *testing.T) {
	router, _, cleanup := setupModuleTest(t)
	defer cleanup()

	t.Run("get recipe with invalid id format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes/invalid-id", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "recipe not found", response["error"])
	})

	t.Run("get non-existing recipe", func(t *testing.T) {
		nonExistingID := primitive.NewObjectID().Hex()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/recipes/"+nonExistingID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "recipe not found", response["error"])
	})

	t.Run("update non-existing recipe", func(t *testing.T) {
		nonExistingID := primitive.NewObjectID().Hex()
		updateBody := map[string]interface{}{
			"title":       "Updated Recipe",
			"description": "This is a valid description",
			"ingredients": []map[string]interface{}{
				{"name": "Test", "amount": 1, "unit": "g"},
			},
			"steps": []map[string]interface{}{
				{"order": 1, "description": "Step", "duration": 1},
			},
			"cookingTime": 10,
			"portions":    2,
			"tags":        []string{"test"},
		}

		bodyJSON, _ := json.Marshal(updateBody)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(
			"PUT",
			"/api/v1/recipes/"+nonExistingID,
			bytes.NewBuffer(bodyJSON),
		)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "recipe not found", response["error"])
	})

	t.Run("delete non-existing recipe", func(t *testing.T) {
		nonExistingID := primitive.NewObjectID().Hex()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/recipes/"+nonExistingID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "recipe not found", response["error"])
	})

	t.Run("create recipe with invalid data - empty title", func(t *testing.T) {
		body := map[string]interface{}{
			"title":       "",
			"description": "Valid description",
			"ingredients": []map[string]interface{}{
				{"name": "Test", "amount": 1, "unit": "g"},
			},
			"steps": []map[string]interface{}{
				{"order": 1, "description": "Step", "duration": 1},
			},
			"cookingTime": 10,
			"portions":    2,
			"tags":        []string{"test"},
		}

		bodyJSON, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/recipes", bytes.NewBuffer(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create recipe with invalid data - no ingredients", func(t *testing.T) {
		body := map[string]interface{}{
			"title":       "Valid Title",
			"description": "Valid description that is long enough",
			"ingredients": []map[string]interface{}{},
			"steps": []map[string]interface{}{
				{"order": 1, "description": "Step", "duration": 1},
			},
			"cookingTime": 10,
			"portions":    2,
			"tags":        []string{"test"},
		}

		bodyJSON, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/recipes", bytes.NewBuffer(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create recipe with invalid json", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(
			"POST",
			"/api/v1/recipes",
			bytes.NewBuffer([]byte(`{invalid json`)),
		)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update recipe with invalid data", func(t *testing.T) {
		// First create a valid recipe
		createBody := map[string]interface{}{
			"title":       "Valid Recipe",
			"description": "This is a valid description for testing",
			"ingredients": []map[string]interface{}{
				{"name": "Test", "amount": 1, "unit": "g"},
			},
			"steps": []map[string]interface{}{
				{"order": 1, "description": "Step", "duration": 1},
			},
			"cookingTime": 10,
			"portions":    2,
			"tags":        []string{"test"},
		}

		bodyJSON, _ := json.Marshal(createBody)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/recipes", bytes.NewBuffer(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		var created handler.RecipeResponse
		err := json.Unmarshal(w.Body.Bytes(), &created)
		require.NoError(t, err)

		// Try to update with invalid data
		updateBody := map[string]interface{}{
			"title":       "",
			"description": "Valid description",
			"ingredients": []map[string]interface{}{
				{"name": "Test", "amount": 1, "unit": "g"},
			},
			"steps": []map[string]interface{}{
				{"order": 1, "description": "Step", "duration": 1},
			},
			"cookingTime": 10,
			"portions":    2,
			"tags":        []string{"test"},
		}

		updateJSON, _ := json.Marshal(updateBody)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("PUT", "/api/v1/recipes/"+created.ID, bytes.NewBuffer(updateJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
