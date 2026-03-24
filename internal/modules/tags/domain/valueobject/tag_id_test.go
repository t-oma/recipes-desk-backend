package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/tags/domain/valueobject"
)

func TestNewTagID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "valid id",
			id:      "abc123",
			wantErr: nil,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: valueobject.ErrTagIDEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewTagID(tt.id)
			if tt.wantErr != nil {
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.Equal(t, tt.id, got.String())
			}
		})
	}
}
