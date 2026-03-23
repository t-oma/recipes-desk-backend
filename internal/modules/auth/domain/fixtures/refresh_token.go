// Package fixtures provides test fixtures for auth domain entities.
package fixtures

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/valueobject"
)

// NewRefreshToken creates a valid refresh token entity for testing.
func NewRefreshToken(t *testing.T, userID string, tokenHash string) *entity.RefreshToken {
	t.Helper()

	id, err := valueobject.NewRefreshTokenID("test-refresh-token-id")
	require.NoError(t, err)

	userIDVO, err := valueobject.NewUserID(userID)
	require.NoError(t, err)

	tokenHashVO := valueobject.TokenHash(tokenHash)

	return entity.NewRefreshToken(id, userIDVO, tokenHashVO)
}

// NewRefreshTokenWithOptions creates a refresh token with custom options.
func NewRefreshTokenWithOptions(t *testing.T, opts ...RefreshTokenOption) *entity.RefreshToken {
	t.Helper()

	options := &refreshTokenOptions{
		id:        "test-refresh-token-id",
		userID:    "test-user-id",
		tokenHash: "test-token-hash",
	}

	for _, opt := range opts {
		opt(options)
	}

	id, err := valueobject.NewRefreshTokenID(options.id)
	require.NoError(t, err)

	userIDVO, err := valueobject.NewUserID(options.userID)
	require.NoError(t, err)

	tokenHashVO := valueobject.TokenHash(options.tokenHash)

	return entity.NewRefreshToken(id, userIDVO, tokenHashVO)
}

// refreshTokenOptions holds options for creating a refresh token.
type refreshTokenOptions struct {
	id        string
	userID    string
	tokenHash string
}

// RefreshTokenOption is a function that configures refreshTokenOptions.
type RefreshTokenOption func(*refreshTokenOptions)

// WithRefreshTokenID sets the ID for the refresh token.
func WithRefreshTokenID(id string) RefreshTokenOption {
	return func(r *refreshTokenOptions) {
		r.id = id
	}
}

// WithRefreshTokenUserID sets the user ID for the refresh token.
func WithRefreshTokenUserID(userID string) RefreshTokenOption {
	return func(r *refreshTokenOptions) {
		r.userID = userID
	}
}

// WithRefreshTokenHash sets the token hash for the refresh token.
func WithRefreshTokenHash(tokenHash string) RefreshTokenOption {
	return func(r *refreshTokenOptions) {
		r.tokenHash = tokenHash
	}
}

// RefreshTokenWithPersistence restores a refresh token from persistence data.
func RefreshTokenWithPersistence(
	t *testing.T,
	token *entity.RefreshToken,
	createdAt, expiresAt time.Time,
) *entity.RefreshToken {
	t.Helper()
	token.RestoreFromPersistence(createdAt, expiresAt)
	return token
}
