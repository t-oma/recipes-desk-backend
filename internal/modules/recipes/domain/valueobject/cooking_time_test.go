package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

func TestNewCookingTime(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		wantErr error
	}{
		{
			name:    "valid time",
			seconds: 60,
			wantErr: nil,
		},
		{
			name:    "invalid time",
			seconds: -1,
			wantErr: valueobject.ErrCookingNegativeTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewCookingTime(tt.seconds)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotNil(t, got)
				require.NotEmpty(t, got.String())
				require.InDelta(t, tt.seconds, got.Duration().Seconds(), 0.01)
			}
		})
	}
}
