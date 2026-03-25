package valueobject_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/tags/domain/valueobject"
)

func TestNewTagName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid name",
			input:   "Italian",
			want:    "Italian",
			wantErr: nil,
		},
		{
			name:    "valid name - min length",
			input:   "AB",
			want:    "AB",
			wantErr: nil,
		},
		{
			name:    "valid name - max length",
			input:   strings.Repeat("a", 50),
			want:    strings.Repeat("a", 50),
			wantErr: nil,
		},
		{
			name:    "trims whitespace",
			input:   "  Italian  ",
			want:    "Italian",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: valueobject.ErrTagNameEmpty,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    "",
			wantErr: valueobject.ErrTagNameEmpty,
		},
		{
			name:    "too short",
			input:   "A",
			want:    "",
			wantErr: valueobject.ErrTagNameTooShort,
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 51),
			want:    "",
			wantErr: valueobject.ErrTagNameTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewTagName(tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.Equal(t, tt.want, got.String())
			}
		})
	}
}
