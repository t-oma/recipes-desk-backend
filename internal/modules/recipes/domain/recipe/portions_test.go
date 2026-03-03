package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/recipe"
)

func TestNewPortions(t *testing.T) {
	tests := []struct {
		name     string
		portions int
		wantErr  error
	}{
		{
			name:     "valid portions",
			portions: recipe.PortionsMax - 1,
			wantErr:  nil,
		},
		{
			name:     "too few portions",
			portions: recipe.PortionsMin - 1,
			wantErr:  recipe.ErrPortionsTooFew,
		},
		{
			name:     "too many portions",
			portions: recipe.PortionsMax + 1,
			wantErr:  recipe.ErrPortionsTooMany,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := recipe.NewPortions(tt.portions)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotNil(t, got)
				require.NotEmpty(t, got.String())
				require.Equal(t, tt.portions, got.Value())
			}
		})
	}
}
