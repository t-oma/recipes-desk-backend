package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/valueobject"
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
			wantErr: valueobject.ErrDescriptionEmpty,
		},
		{
			name:    "only whitespaces",
			desc:    "         ",
			wantErr: valueobject.ErrDescriptionEmpty,
		},
		{
			name:    "description too short",
			desc:    stringutil.RandomString(valueobject.DescriptionMinLength - 1),
			wantErr: valueobject.ErrDescriptionTooShort,
		},
		{
			name:    "description too long",
			desc:    stringutil.RandomString(valueobject.DescriptionMaxLength + 1),
			wantErr: valueobject.ErrDescriptionTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewDescription(tt.desc)
			if tt.wantErr != nil {
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				require.NoError(t, gotErr)
				require.NotEmpty(t, got.String())
			}
		})
	}
}
