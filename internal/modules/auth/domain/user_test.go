package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"recipes-desk/internal/modules/auth/domain"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		wantErr error
	}{
		{
			name: "valid user",
			user: domain.User{ //nolint:exhaustruct // test struct
				ID:        "507f1f77bcf86cd799439011",
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			wantErr: nil,
		},
		{
			name: "empty email",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			wantErr: domain.ErrEmptyEmail,
		},
		{
			name: "email too short",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "a",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			wantErr: domain.ErrInvalidEmailLength,
		},
		{
			name: "email too long",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     string(make([]byte, 255)),
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			wantErr: domain.ErrInvalidEmailLength,
		},
		{
			name: "empty password",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "",
			},
			wantErr: domain.ErrEmptyPassword,
		},
		{
			name: "password too short",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "short",
			},
			wantErr: domain.ErrPasswordTooShort,
		},
		{
			name: "password too long",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  string(make([]byte, 73)),
			},
			wantErr: domain.ErrPasswordTooLong,
		},
		{
			name: "empty first name",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "Doe",
				Password:  "password123",
			},
			wantErr: domain.ErrEmptyFirstName,
		},
		{
			name: "first name too short",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: "J",
				LastName:  "Doe",
				Password:  "password123",
			},
			wantErr: domain.ErrInvalidFirstNameLength,
		},
		{
			name: "first name too long",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: string(make([]byte, 51)),
				LastName:  "Doe",
				Password:  "password123",
			},
			wantErr: domain.ErrInvalidFirstNameLength,
		},
		{
			name: "empty last name",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "",
				Password:  "password123",
			},
			wantErr: domain.ErrEmptyLastName,
		},
		{
			name: "last name too short",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "D",
				Password:  "password123",
			},
			wantErr: domain.ErrInvalidLastNameLength,
		},
		{
			name: "last name too long",
			user: domain.User{ //nolint:exhaustruct // test struct
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  string(make([]byte, 51)),
				Password:  "password123",
			},
			wantErr: domain.ErrInvalidLastNameLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
