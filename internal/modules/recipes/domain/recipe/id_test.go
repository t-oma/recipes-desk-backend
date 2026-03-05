package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/recipe"
	"recipes-desk/pkg/stringutil"
)

func TestNewRecipeID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "valid id",
			id:      stringutil.RandomString(10),
			wantErr: nil,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: recipe.ErrRecipeIDEmpty,
		},
		{
			name:    "only whitespaces",
			id:      "         ",
			wantErr: recipe.ErrRecipeIDEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := recipe.NewEntityID(tt.id)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotEmpty(t, got.String())
			}
		})
	}
}
