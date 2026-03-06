package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

func TestNewAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr error
	}{
		{
			name:    "valid amount",
			amount:  1.0,
			wantErr: nil,
		},
		{
			name:    "too few amount",
			amount:  valueobject.AmountMin - 1,
			wantErr: valueobject.ErrAmountToFew,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewAmount(tt.amount)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotNil(t, got)
				require.NotEmpty(t, got.String())
				require.InEpsilon(t, tt.amount, got.Value(), 0.01)
			}
		})
	}
}
