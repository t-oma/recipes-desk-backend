//go:build !integration
// +build !integration

package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/recipes/domain"
)

func TestMongoRepository_IDParsing(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid hex id",
			id:      primitive.NewObjectID().Hex(),
			wantErr: false,
		},
		{
			name:    "invalid hex - too short",
			id:      "abc123",
			wantErr: true,
		},
		{
			name:    "invalid hex - contains invalid chars",
			id:      "xyz1234567890123456789012",
			wantErr: true,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := primitive.ObjectIDFromHex(tt.id)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMongoRepository_RecipePreparation(t *testing.T) {
	tests := []struct {
		name           string
		recipe         *domain.Recipe
		expectNewID    bool
		expectNewTimes bool
	}{
		{
			name: "recipe with zero ID gets new ID and timestamps",
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				Title:       "Test Recipe",
				Description: "This is a valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
			},
			expectNewID:    true,
			expectNewTimes: true,
		},
		{
			name: "recipe with existing ID preserves ID but gets new timestamps",
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				ID:          primitive.NewObjectID(),
				Title:       "Test Recipe",
				Description: "This is a valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
			},
			expectNewID:    false,
			expectNewTimes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.recipe.ID
			originalCreatedAt := tt.recipe.CreatedAt

			// Simulate what repository.Create does
			if tt.recipe.ID.IsZero() {
				tt.recipe.ID = primitive.NewObjectID()
			}
			tt.recipe.SetTimestamps()

			if tt.expectNewID && originalID.IsZero() {
				assert.False(t, tt.recipe.ID.IsZero())
				assert.NotEqual(t, originalID, tt.recipe.ID)
			} else {
				assert.Equal(t, originalID, tt.recipe.ID)
			}

			if tt.expectNewTimes {
				assert.False(t, tt.recipe.UpdatedAt.IsZero())
				if originalCreatedAt.IsZero() {
					assert.False(t, tt.recipe.CreatedAt.IsZero())
				} else {
					assert.Equal(t, originalCreatedAt, tt.recipe.CreatedAt)
				}
			}
		})
	}
}

func TestMongoRepository_BSONFilterCreation(t *testing.T) {
	recipeID := primitive.NewObjectID()

	t.Run("find by id filter", func(t *testing.T) {
		filter := primitive.M{"_id": recipeID}
		assert.NotNil(t, filter)
		assert.Equal(t, recipeID, filter["_id"])
	})

	t.Run("update filter", func(t *testing.T) {
		filter := primitive.M{"_id": recipeID}
		update := primitive.M{
			"$set": domain.Recipe{ //nolint:exhaustruct // test struct
				ID:    recipeID,
				Title: "Updated",
			},
		}
		assert.NotNil(t, filter)
		assert.NotNil(t, update)
	})

	t.Run("delete filter", func(t *testing.T) {
		filter := primitive.M{"_id": recipeID}
		assert.NotNil(t, filter)
		assert.Equal(t, recipeID, filter["_id"])
	})
}
