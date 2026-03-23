package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

func TestNewUnit(t *testing.T) {
	tests := []struct {
		name     string
		unitName string
		wantErr  error
	}{
		{
			name:     "valid unit",
			unitName: valueobject.UnitCup,
			wantErr:  nil,
		},
		{
			name:     "empty unit",
			unitName: "",
			wantErr:  valueobject.ErrUnitEmptyName,
		},
		{
			name:     "unit too short",
			unitName: "not-a-unit",
			wantErr:  valueobject.ErrUnitUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewUnit(tt.unitName)
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
