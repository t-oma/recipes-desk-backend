package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/recipe"
	"recipes-desk/pkg/stringutil"
)

func TestNewDescription(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		wantErr error
	}{
		{
			name:    "valid description",
			desc:    "Classic Italian pasta dish with eggs and cheese",
			wantErr: nil,
		},
		{
			name:    "empty description",
			desc:    "",
			wantErr: recipe.ErrDescriptionEmpty,
		},
		{
			name:    "only whitespaces",
			desc:    "         ",
			wantErr: recipe.ErrDescriptionEmpty,
		},
		{
			name:    "description too short",
			desc:    stringutil.RandomString(recipe.DescriptionMinLength - 1),
			wantErr: recipe.ErrDescriptionTooShort,
		},
		{
			name:    "description too long",
			desc:    stringutil.RandomString(recipe.DescriptionMaxLength + 1),
			wantErr: recipe.ErrDescriptionTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := recipe.NewDescription(tt.desc)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotEmpty(t, got.String())
				require.NotEmpty(t, got.Value())
			}
		})
	}
}
