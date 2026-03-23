package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

func TestNewPortions(t *testing.T) {
	tests := []struct {
		name     string
		portions int
		wantErr  error
	}{
		{
			name:     "valid portions",
			portions: valueobject.PortionsMax - 1,
			wantErr:  nil,
		},
		{
			name:     "too few portions",
			portions: valueobject.PortionsMin - 1,
			wantErr:  valueobject.ErrPortionsTooFew,
		},
		{
			name:     "too many portions",
			portions: valueobject.PortionsMax + 1,
			wantErr:  valueobject.ErrPortionsTooMany,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewPortions(tt.portions)
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
