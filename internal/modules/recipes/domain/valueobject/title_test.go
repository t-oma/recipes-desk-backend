package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/valueobject"
	"recipes-desk/pkg/stringutil"
)

func TestNewTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr error
	}{
		{
			name:    "valid title",
			title:   "Pasta Carbonara",
			wantErr: nil,
		},
		{
			name:    "empty title",
			title:   "",
			wantErr: valueobject.ErrTitleEmpty,
		},
		{
			name:    "title too short",
			title:   stringutil.RandomString(valueobject.TitleMinLength - 1),
			wantErr: valueobject.ErrTitleTooShort,
		},
		{
			name:    "title too long",
			title:   stringutil.RandomString(valueobject.TitleMaxLength + 1),
			wantErr: valueobject.ErrTitleTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := valueobject.NewTitle(tt.title)
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
