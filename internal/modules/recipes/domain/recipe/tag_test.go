package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/recipe"
	"recipes-desk/pkg/stringutil"
)

func TestNewTag(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		wantErr error
	}{
		{
			name:    "valid tag",
			tagName: "italian",
			wantErr: nil,
		},
		{
			name:    "empty tag",
			tagName: "",
			wantErr: recipe.ErrTagEmptyName,
		},
		{
			name:    "tag too long",
			tagName: stringutil.RandomString(recipe.TagMaxLength + 1),
			wantErr: recipe.ErrTagNameTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := recipe.NewTag(tt.tagName)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotEmpty(t, got.String())
				require.NotEmpty(t, got.Name())
			}
		})
	}
}
