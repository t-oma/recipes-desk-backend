package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
			recipe: domain.Recipe{ //nolint:exhaustruct // test struct
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
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRecipe_Validate_WrappedErrors(t *testing.T) {
	// Test that validation errors are properly wrapped with ErrValidation
	recipe := domain.Recipe{ //nolint:exhaustruct // test struct
		Title:       "",
		Description: "Short",
	}

	err := recipe.Validate()

	// Check that we get the specific error
	require.ErrorIs(t, err, domain.ErrEmptyTitle)

	// Check that the error message contains validation info
	assert.Contains(t, err.Error(), "validation error")
	assert.Contains(t, err.Error(), "title")
}
