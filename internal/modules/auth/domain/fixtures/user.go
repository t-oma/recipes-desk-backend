// Package fixtures provides test fixtures for auth domain entities.
package fixtures

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/valueobject"
)

// NewUser creates a valid user entity for testing.
func NewUser(t *testing.T, email string) *entity.User {
	t.Helper()

	id, err := valueobject.NewUserID(primitive.NewObjectID().Hex())
	require.NoError(t, err)

	emailVO, err := valueobject.NewEmail(email)
	require.NoError(t, err)

	firstName, err := valueobject.NewFirstName("John")
	require.NoError(t, err)

	lastName, err := valueobject.NewLastName("Doe")
	require.NoError(t, err)

	passwordHash := valueobject.PasswordHash("hashedpassword123")

	user := entity.NewUser(id, emailVO, firstName, lastName, passwordHash)

	return user
}

// NewUserWithName creates a user with specific first and last name.
func NewUserWithName(t *testing.T, email, firstNameStr, lastNameStr string) *entity.User {
	t.Helper()

	id, err := valueobject.NewUserID(primitive.NewObjectID().Hex())
	require.NoError(t, err)

	emailVO, err := valueobject.NewEmail(email)
	require.NoError(t, err)

	firstName, err := valueobject.NewFirstName(firstNameStr)
	require.NoError(t, err)

	lastName, err := valueobject.NewLastName(lastNameStr)
	require.NoError(t, err)

	passwordHash := valueobject.PasswordHash("hashedpassword123")

	user := entity.NewUser(id, emailVO, firstName, lastName, passwordHash)

	return user
}

// NewUserWithOptions creates a user with custom options.
func NewUserWithOptions(t *testing.T, opts ...UserOption) *entity.User {
	t.Helper()

	options := &userOptions{
		email:     "test@example.com",
		firstName: "John",
		lastName:  "Doe",
		password:  "hashedpassword123",
	}

	for _, opt := range opts {
		opt(options)
	}

	id, err := valueobject.NewUserID(primitive.NewObjectID().Hex())
	require.NoError(t, err)

	emailVO, err := valueobject.NewEmail(options.email)
	require.NoError(t, err)

	firstName, err := valueobject.NewFirstName(options.firstName)
	require.NoError(t, err)

	lastName, err := valueobject.NewLastName(options.lastName)
	require.NoError(t, err)

	passwordHash := valueobject.PasswordHash(options.password)

	user := entity.NewUser(id, emailVO, firstName, lastName, passwordHash)

	return user
}

// userOptions holds options for creating a user.
type userOptions struct {
	email     string
	firstName string
	lastName  string
	password  string
}

// UserOption is a function that configures userOptions.
type UserOption func(*userOptions)

// WithEmail sets the email for the user.
func WithEmail(email string) UserOption {
	return func(u *userOptions) {
		u.email = email
	}
}

// WithFirstName sets the first name for the user.
func WithFirstName(firstName string) UserOption {
	return func(u *userOptions) {
		u.firstName = firstName
	}
}

// WithLastName sets the last name for the user.
func WithLastName(lastName string) UserOption {
	return func(u *userOptions) {
		u.lastName = lastName
	}
}

// WithPassword sets the password hash for the user.
func WithPassword(password string) UserOption {
	return func(u *userOptions) {
		u.password = password
	}
}
