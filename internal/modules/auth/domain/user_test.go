package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
			user: domain.User{ //nolint:exhaustruct // test struct //nolint:exhaustruct // test struct - only validation fields needed
				ID:        primitive.NewObjectID(),
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

func TestUser_SetTimestamps(t *testing.T) {
	t.Run("sets timestamps on new user", func(t *testing.T) {
		user := domain.User{ //nolint:exhaustruct // test struct
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Password:  "password123",
		}

		before := time.Now()
		user.SetTimestamps()
		after := time.Now()

		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.PasswordUpdatedAt.IsZero())
		assert.True(t, user.CreatedAt.After(before) || user.CreatedAt.Equal(before))
		assert.True(t, user.CreatedAt.Before(after) || user.CreatedAt.Equal(after))
	})

	t.Run("preserves existing CreatedAt", func(t *testing.T) {
		existingTime := time.Now().Add(-24 * time.Hour)
		user := domain.User{ //nolint:exhaustruct // test struct
			Email:             "test@example.com",
			FirstName:         "John",
			LastName:          "Doe",
			Password:          "password123",
			CreatedAt:         existingTime,
			PasswordUpdatedAt: time.Time{},
		}

		user.SetTimestamps()

		assert.Equal(t, existingTime, user.CreatedAt)
		assert.False(t, user.PasswordUpdatedAt.IsZero())
	})

	t.Run("preserves existing PasswordUpdatedAt", func(t *testing.T) {
		oldPasswordTime := time.Now().Add(-24 * time.Hour)
		user := domain.User{ //nolint:exhaustruct // test struct
			Email:             "test@example.com",
			FirstName:         "John",
			LastName:          "Doe",
			Password:          "newpassword123",
			CreatedAt:         time.Now().Add(-48 * time.Hour),
			PasswordUpdatedAt: oldPasswordTime,
		}

		user.SetTimestamps()

		// PasswordUpdatedAt should be preserved if already set
		assert.Equal(t, oldPasswordTime, user.PasswordUpdatedAt)
	})
}
