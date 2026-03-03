package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/recipe"
	"recipes-desk/pkg/stringutil"
)

func TestNewStep(t *testing.T) {
	tests := []struct {
		name        string
		order       int
		description string
		durationMin int
		wantErr     error
	}{
		{
			name:        "valid step",
			order:       1,
			description: "Mix ingredients",
			durationMin: 5,
			wantErr:     nil,
		},
		{
			name:        "step order too low",
			order:       0,
			description: "Mix ingredients",
			durationMin: 5,
			wantErr:     recipe.ErrInvalidStepOrder,
		},
		{
			name:        "step description too short",
			order:       1,
			description: stringutil.RandomString(recipe.StepMinDescriptionLength - 1),
			durationMin: 5,
			wantErr:     recipe.ErrStepDescriptionTooShort,
		},
		{
			name:        "step description too long",
			order:       1,
			description: stringutil.RandomString(recipe.StepMaxDescriptionLength + 1),
			durationMin: 5,
			wantErr:     recipe.ErrStepDescriptionTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := recipe.NewStep(tt.order, tt.description, tt.durationMin)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotEmpty(t, got.String())
				require.NotEmpty(t, got.Description())
				require.NotEmpty(t, got.Duration())
				require.Equal(t, tt.order, got.Order())
			}
		})
	}
}
