package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/recipes/domain"
)

func TestRecipe_Validate(t *testing.T) {
	tests := []struct {
		name    string
		recipe  domain.Recipe
		wantErr error
	}{
		{
			name: "valid recipe",
			recipe: domain.Recipe{
				Title:       "Pasta Carbonara",
				Description: "Classic Italian pasta dish with eggs and cheese",
				Ingredients: []domain.Ingredient{
					{Name: "Spaghetti", Amount: 400, Unit: "g"},
					{Name: "Eggs", Amount: 4, Unit: "pcs"},
				},
				Steps: []domain.Step{
					{Order: 1, Description: "Boil pasta", Duration: 10},
					{Order: 2, Description: "Mix eggs with cheese", Duration: 5},
				},
				CookingTime: 20,
				Portions:    4,
				Tags:        []string{"italian", "pasta"},
			},
			wantErr: nil,
		},
		{
			name: "empty title",
			recipe: domain.Recipe{
				Title:       "",
				Description: "Some description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrEmptyTitle,
		},
		{
			name: "title too short",
			recipe: domain.Recipe{
				Title:       "Ab",
				Description: "Some description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrInvalidTitleLength,
		},
		{
			name: "title too long",
			recipe: domain.Recipe{
				Title:       string(make([]byte, 201)), // 201 characters
				Description: "Some description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrInvalidTitleLength,
		},
		{
			name: "empty description",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrEmptyDescription,
		},
		{
			name: "description too short",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "Short",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrInvalidDescriptionLength,
		},
		{
			name: "no ingredients",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "This is a valid description that is long enough",
				Ingredients: []domain.Ingredient{},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrNoIngredients,
		},
		{
			name: "no steps",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "This is a valid description that is long enough",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrNoSteps,
		},
		{
			name: "invalid cooking time - zero",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "This is a valid description that is long enough",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 0,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrInvalidCookingTime,
		},
		{
			name: "invalid cooking time - negative",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "This is a valid description that is long enough",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: -5,
				Portions:    2,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrInvalidCookingTime,
		},
		{
			name: "invalid portions - zero",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "This is a valid description that is long enough",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    0,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrInvalidPortions,
		},
		{
			name: "invalid portions - too many",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "This is a valid description that is long enough",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    101,
				Tags:        []string{"tag"},
			},
			wantErr: domain.ErrInvalidPortions,
		},
		{
			name: "no tags",
			recipe: domain.Recipe{
				Title:       "Valid Title",
				Description: "This is a valid description that is long enough",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{},
			},
			wantErr: domain.ErrNoTags,
		},
		{
			name: "valid recipe - boundary values",
			recipe: domain.Recipe{
				Title:       "ABC",        // minimum 3 chars
				Description: "1234567890", // minimum 10 chars
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 0}},
				CookingTime: 1, // minimum > 0
				Portions:    1, // minimum
				Tags:        []string{"tag"},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.recipe.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecipe_SetTimestamps(t *testing.T) {
	tests := []struct {
		name          string
		recipe        domain.Recipe
		wantCreatedAt bool
		wantUpdatedAt bool
	}{
		{
			name: "new recipe - sets both timestamps",
			recipe: domain.Recipe{
				ID: primitive.NewObjectID(),
			},
			wantCreatedAt: true,
			wantUpdatedAt: true,
		},
		{
			name: "existing recipe - updates only updatedAt",
			recipe: domain.Recipe{
				ID:        primitive.NewObjectID(),
				CreatedAt: time.Now().Add(-time.Hour),
			},
			wantCreatedAt: true,
			wantUpdatedAt: true,
		},
		{
			name: "recipe with zero values - sets both",
			recipe: domain.Recipe{
				ID:        primitive.NilObjectID,
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			wantCreatedAt: true,
			wantUpdatedAt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCreatedAt := tt.recipe.CreatedAt

			tt.recipe.SetTimestamps()

			// UpdatedAt should always be set to now
			assert.False(t, tt.recipe.UpdatedAt.IsZero(), "UpdatedAt should be set")
			assert.WithinDuration(t, time.Now(), tt.recipe.UpdatedAt, time.Second)

			if tt.wantCreatedAt {
				if originalCreatedAt.IsZero() {
					// If it was zero, it should be set to now
					assert.False(t, tt.recipe.CreatedAt.IsZero(), "CreatedAt should be set")
					assert.WithinDuration(t, time.Now(), tt.recipe.CreatedAt, time.Second)
				} else {
					// If it had a value, it should be preserved
					assert.Equal(
						t,
						originalCreatedAt,
						tt.recipe.CreatedAt,
						"CreatedAt should be preserved",
					)
				}
			}
		})
	}
}

func TestRecipe_SetTimestamps_PreservesCreatedAt(t *testing.T) {
	// Specific test for the preservation of CreatedAt
	originalTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	recipe := domain.Recipe{
		ID:        primitive.NewObjectID(),
		CreatedAt: originalTime,
		UpdatedAt: originalTime,
	}

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	recipe.SetTimestamps()

	// CreatedAt should be preserved
	assert.Equal(t, originalTime, recipe.CreatedAt, "CreatedAt should be preserved")

	// UpdatedAt should be updated
	assert.True(t, recipe.UpdatedAt.After(originalTime), "UpdatedAt should be after original time")
}

func TestRecipe_Validate_WrappedErrors(t *testing.T) {
	// Test that validation errors are properly wrapped with ErrValidation
	recipe := domain.Recipe{
		Title:       "",
		Description: "Short",
	}

	err := recipe.Validate()

	// Check that we get the specific error
	assert.ErrorIs(t, err, domain.ErrEmptyTitle)

	// Check that the error message contains validation info
	assert.Contains(t, err.Error(), "validation error")
	assert.Contains(t, err.Error(), "title")
}
