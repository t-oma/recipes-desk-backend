package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/tags/domain/valueobject"
)

func TestNewTagSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "lowercases uppercase",
			input:   "Italian",
			want:    "italian",
			wantErr: nil,
		},
		{
			name:    "lowercases mixed case",
			input:   "Gluten Free",
			want:    "gluten-free",
			wantErr: nil,
		},
		{
			name:    "preserves already lowercase",
			input:   "breakfast",
			want:    "breakfast",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: valueobject.ErrTagNameEmpty,
		},
		{
			name:    "too short",
			input:   "A",
			want:    "",
			wantErr: valueobject.ErrTagNameTooShort,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewTagSlug(tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.Equal(t, tt.want, got.String())
			}
		})
	}
}
