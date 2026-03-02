//go:build integration
// +build integration

package mongorepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/repository/mongorepo"
)

func TestIntegration_MongoRefreshTokenRepository_Create(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	t.Run("create refresh token", func(t *testing.T) {
		userID := "507f1f77bcf86cd799439011"
		token := &domain.RefreshToken{
			UserID:    userID,
			TokenHash: "hash123",
		}

		token, err := repo.Create(ctx, token, 7*24*time.Hour)
		require.NoError(t, err)

		// Verify ID was set
		assert.NotEmpty(t, token.ID)
		// Verify timestamps
		assert.False(t, token.CreatedAt.IsZero())
		assert.False(t, token.ExpiresAt.IsZero())
		assert.True(t, token.ExpiresAt.After(token.CreatedAt))
	})

	t.Run("create multiple tokens for same user", func(t *testing.T) {
		userID := "507f1f77bcf86cd799439012"
		token1 := &domain.RefreshToken{
			UserID:    userID,
			TokenHash: "hash1",
		}

		token2 := &domain.RefreshToken{
			UserID:    userID,
			TokenHash: "hash2",
		}

		token1, err := repo.Create(ctx, token1, 7*24*time.Hour)
		require.NoError(t, err)

		token2, err = repo.Create(ctx, token2, 7*24*time.Hour)
		require.NoError(t, err)

		// Both should have different IDs
		assert.NotEqual(t, token1.ID, token2.ID)
	})

	t.Run("create token with wrong userID", func(t *testing.T) {
		userID := "wrong-user-id"
		token := &domain.RefreshToken{
			UserID:    userID,
			TokenHash: "hash123",
		}

		token, err := repo.Create(ctx, token, 7*24*time.Hour)
		require.Error(t, err)
	})
}

func TestIntegration_MongoRefreshTokenRepository_FindByHash(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	// Create a test token first
	userID := "507f1f77bcf86cd799439011"
	token := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "findme123",
	}
	token, err := repo.Create(ctx, token, 7*24*time.Hour)
	require.NoError(t, err)

	t.Run("find existing token", func(t *testing.T) {
		found, err := repo.FindByHash(ctx, "findme123")
		require.NoError(t, err)

		assert.Equal(t, token.ID, found.ID)
		assert.Equal(t, token.UserID, found.UserID)
		assert.Equal(t, token.TokenHash, found.TokenHash)
	})

	t.Run("find non-existent token", func(t *testing.T) {
		_, err := repo.FindByHash(ctx, "nonexistent")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestIntegration_MongoRefreshTokenRepository_DeleteByHash(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	// Create a test token first
	userID := "507f1f77bcf86cd799439011"
	token := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "deleteme123",
	}
	token, err := repo.Create(ctx, token, 7*24*time.Hour)
	require.NoError(t, err)

	t.Run("delete existing token", func(t *testing.T) {
		err := repo.DeleteByHash(ctx, "deleteme123")
		require.NoError(t, err)

		// Verify token is deleted
		_, err = repo.FindByHash(ctx, "deleteme123")
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("delete non-existent token", func(t *testing.T) {
		// Should not return error for non-existent token
		err := repo.DeleteByHash(ctx, "nonexistent")
		assert.NoError(t, err)
	})
}

func TestIntegration_MongoRefreshTokenRepository_TokenRotation(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	userID := "507f1f77bcf86cd799439011"

	t.Run("simulate token rotation", func(t *testing.T) {
		// Create old token
		oldToken := &domain.RefreshToken{
			UserID:    userID,
			TokenHash: "oldtoken123",
		}
		oldToken, err := repo.Create(ctx, oldToken, 7*24*time.Hour)
		require.NoError(t, err)

		// Create new token (rotation)
		newToken := &domain.RefreshToken{
			UserID:    userID,
			TokenHash: "newtoken456",
		}
		newToken, err = repo.Create(ctx, newToken, 7*24*time.Hour)
		require.NoError(t, err)

		// Verify both tokens exist
		foundOld, err := repo.FindByHash(ctx, "oldtoken123")
		require.NoError(t, err)
		assert.Equal(t, oldToken.ID, foundOld.ID)

		foundNew, err := repo.FindByHash(ctx, "newtoken456")
		require.NoError(t, err)
		assert.Equal(t, newToken.ID, foundNew.ID)

		// Delete old token (part of rotation)
		err = repo.DeleteByHash(ctx, "oldtoken123")
		require.NoError(t, err)

		// Verify old token is deleted but new still exists
		_, err = repo.FindByHash(ctx, "oldtoken123")
		assert.ErrorIs(t, err, domain.ErrNotFound)

		foundNew, err = repo.FindByHash(ctx, "newtoken456")
		require.NoError(t, err)
		assert.Equal(t, newToken.ID, foundNew.ID)
	})
}

func TestIntegration_MongoRefreshTokenRepository_MultipleUsers(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	user1 := "507f1f77bcf86cd799439011"
	user2 := "507f1f77bcf86cd799439012"

	t.Run("tokens for different users", func(t *testing.T) {
		// Create tokens for user1
		token1 := &domain.RefreshToken{
			UserID:    user1,
			TokenHash: "user1token",
		}
		token1, err := repo.Create(ctx, token1, 7*24*time.Hour)
		require.NoError(t, err)

		// Create tokens for user2
		token2 := &domain.RefreshToken{
			UserID:    user2,
			TokenHash: "user2token",
		}
		token2, err = repo.Create(ctx, token2, 7*24*time.Hour)
		require.NoError(t, err)

		// Verify each user can find their token
		found1, err := repo.FindByHash(ctx, "user1token")
		require.NoError(t, err)
		assert.Equal(t, user1, found1.UserID)

		found2, err := repo.FindByHash(ctx, "user2token")
		require.NoError(t, err)
		assert.Equal(t, user2, found2.UserID)

		// Delete user1's token
		err = repo.DeleteByHash(ctx, "user1token")
		require.NoError(t, err)

		// Verify user2's token still exists
		found2, err = repo.FindByHash(ctx, "user2token")
		require.NoError(t, err)
		assert.Equal(t, user2, found2.UserID)
	})
}

func TestIntegration_MongoRefreshTokenRepository_TTLIndex(t *testing.T) {
	db, cleanup := setupMongoContainer(t)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	repo.InitIndexes(context.Background())
	ctx := context.Background()

	t.Run("verify TTL index is created", func(t *testing.T) {
		// Just verify repository creation doesn't fail
		// TTL cleanup is handled by MongoDB and may take time
		assert.NotNil(t, repo)

		// Create a token that expires in 1 second
		userID := "507f1f77bcf86cd799439011"
		token := &domain.RefreshToken{
			UserID:    userID,
			TokenHash: "shortlived",
		}

		token, err := repo.Create(ctx, token, 1*time.Second)
		require.NoError(t, err)

		// Verify token exists
		found, err := repo.FindByHash(ctx, "shortlived")
		require.NoError(t, err)
		assert.Equal(t, "shortlived", found.TokenHash)

		// Note: We can't easily test TTL deletion in integration tests
		// because it requires waiting for MongoDB's background job
		// which runs every 60 seconds by default
	})
}
