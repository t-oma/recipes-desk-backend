package recipe_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/recipe"
)

func TestNewIngredient(t *testing.T) {
	tests := []struct {
		name           string
		ingredientName string
		amount         float64
		unit           string
		wantErr        error
	}{
		{
			name:           "valid ingredient",
			ingredientName: "Flour",
			amount:         500.0,
			unit:           recipe.UnitGram,
			wantErr:        nil,
		},
		{
			name:           "empty ingredient name",
			ingredientName: "",
			amount:         500.0,
			unit:           recipe.UnitGram,
			wantErr:        recipe.ErrIngredientEmptyName,
		},
		{
			name:           "invalid amount",
			ingredientName: "Flour",
			amount:         -1.0,
			unit:           recipe.UnitGram,
			wantErr:        recipe.ErrAmountToFew,
		},
		{
			name:           "invalid unit",
			ingredientName: "Flour",
			amount:         500.0,
			unit:           "not-a-unit",
			wantErr:        recipe.ErrUnitUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := recipe.NewIngredient(tt.ingredientName, tt.amount, tt.unit)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotNil(t, got)
				require.Equal(t, tt.ingredientName, got.Name())
				require.InDelta(t, tt.amount, got.Amount().Value(), 0.01)
				require.Equal(t, tt.unit, got.Unit().Name())

				require.Contains(t, got.String(), tt.ingredientName)
				require.Contains(t, got.String(), fmt.Sprintf("%.2f", tt.amount))
				require.Contains(t, got.String(), tt.unit)
			}
		})
	}
}

func TestIngredient_Equals(t *testing.T) {
	ingredient1, err := recipe.NewIngredient("Flour", 500, recipe.UnitGram)
	require.NoError(t, err)

	ingredient2, err := recipe.NewIngredient("Flour", 500, recipe.UnitGram)
	require.NoError(t, err)

	ingredient3, err := recipe.NewIngredient("Eggs", 3, recipe.UnitPcs)
	require.NoError(t, err)

	require.False(t, ingredient1.Equals(nil))
	require.True(t, ingredient1.Equals(&ingredient2))
	require.False(t, ingredient1.Equals(&ingredient3))
}
