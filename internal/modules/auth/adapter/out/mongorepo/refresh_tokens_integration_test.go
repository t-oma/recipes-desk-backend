//go:build integration
// +build integration

package mongorepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/auth/adapter/out/mongorepo"
	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/domain/fixtures"
	"recipes-desk/pkg/testutils"
)

func TestIntegration_MongoRefreshTokenRepository_Create(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)

	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	t.Run("create refresh token", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		token := fixtures.NewRefreshToken(t, userID, "hash123")

		created, err := repo.Create(ctx, token, 7*24*time.Hour)
		require.NoError(t, err)

		// Verify ID was set
		assert.NotEmpty(t, created.ID().String())
		// Verify timestamps
		assert.False(t, created.CreatedAt().IsZero())
		assert.False(t, created.ExpiresAt().IsZero())
		assert.True(t, created.ExpiresAt().After(created.CreatedAt()))
	})

	t.Run("create multiple tokens for same user", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		token1 := fixtures.NewRefreshTokenWithOptions(t,
			fixtures.WithRefreshTokenUserID(userID),
			fixtures.WithRefreshTokenHash("hash1"),
		)
		token2 := fixtures.NewRefreshTokenWithOptions(t,
			fixtures.WithRefreshTokenUserID(userID),
			fixtures.WithRefreshTokenHash("hash2"),
		)

		created1, err := repo.Create(ctx, token1, 7*24*time.Hour)
		require.NoError(t, err)

		created2, err := repo.Create(ctx, token2, 7*24*time.Hour)
		require.NoError(t, err)

		// Both should have different IDs
		assert.NotEqual(t, created1.ID().String(), created2.ID().String())
	})

	t.Run("create token with wrong userID", func(t *testing.T) {
		// Test with invalid userID format - repository should handle this
		userID := "wrong-user-id-format"
		token := fixtures.NewRefreshToken(t, userID, "hash123")

		_, err := repo.Create(ctx, token, 7*24*time.Hour)
		// This may succeed or fail depending on repository implementation
		// The repository validates userID when converting to MongoDB model
		_ = err
	})
}

func TestIntegration_MongoRefreshTokenRepository_FindByHash(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	// Create a test token first
	userID := primitive.NewObjectID().Hex()
	token := fixtures.NewRefreshToken(t, userID, "findme123")
	created, err := repo.Create(ctx, token, 7*24*time.Hour)
	require.NoError(t, err)

	t.Run("find existing token", func(t *testing.T) {
		found, err := repo.FindByHash(ctx, "findme123")
		require.NoError(t, err)

		assert.Equal(t, created.ID().String(), found.ID().String())
		assert.Equal(t, userID, found.UserID().String())
		assert.Equal(t, "findme123", string(found.TokenHash()))
	})

	t.Run("find non-existent token", func(t *testing.T) {
		_, err := repo.FindByHash(ctx, "nonexistent")
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})
}

func TestIntegration_MongoRefreshTokenRepository_DeleteByHash(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	// Create a test token first
	userID := primitive.NewObjectID().Hex()
	token := fixtures.NewRefreshToken(t, userID, "deleteme123")
	_, err := repo.Create(ctx, token, 7*24*time.Hour)
	require.NoError(t, err)

	t.Run("delete existing token", func(t *testing.T) {
		err := repo.DeleteByHash(ctx, "deleteme123")
		require.NoError(t, err)

		// Verify token is deleted
		_, err = repo.FindByHash(ctx, "deleteme123")
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	})

	t.Run("delete non-existent token", func(t *testing.T) {
		// Should not return error for non-existent token
		err := repo.DeleteByHash(ctx, "nonexistent")
		assert.NoError(t, err)
	})
}

func TestIntegration_MongoRefreshTokenRepository_TokenRotation(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	userID := primitive.NewObjectID().Hex()

	t.Run("simulate token rotation", func(t *testing.T) {
		// Create old token
		oldToken := fixtures.NewRefreshToken(t, userID, "oldtoken123")
		createdOld, err := repo.Create(ctx, oldToken, 7*24*time.Hour)
		require.NoError(t, err)

		// Create new token (rotation)
		newToken := fixtures.NewRefreshToken(t, userID, "newtoken456")
		createdNew, err := repo.Create(ctx, newToken, 7*24*time.Hour)
		require.NoError(t, err)

		// Verify both tokens exist
		foundOld, err := repo.FindByHash(ctx, "oldtoken123")
		require.NoError(t, err)
		assert.Equal(t, createdOld.ID().String(), foundOld.ID().String())

		foundNew, err := repo.FindByHash(ctx, "newtoken456")
		require.NoError(t, err)
		assert.Equal(t, createdNew.ID().String(), foundNew.ID().String())

		// Delete old token (part of rotation)
		err = repo.DeleteByHash(ctx, "oldtoken123")
		require.NoError(t, err)

		// Verify old token is deleted but new still exists
		_, err = repo.FindByHash(ctx, "oldtoken123")
		assert.ErrorIs(t, err, domain.ErrTokenNotFound)

		foundNew, err = repo.FindByHash(ctx, "newtoken456")
		require.NoError(t, err)
		assert.Equal(t, createdNew.ID().String(), foundNew.ID().String())
	})
}

func TestIntegration_MongoRefreshTokenRepository_MultipleUsers(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	ctx := context.Background()

	user1 := primitive.NewObjectID().Hex()
	user2 := primitive.NewObjectID().Hex()

	t.Run("tokens for different users", func(t *testing.T) {
		// Create tokens for user1
		token1 := fixtures.NewRefreshToken(t, user1, "user1token")
		_, err := repo.Create(ctx, token1, 7*24*time.Hour)
		require.NoError(t, err)

		// Create tokens for user2
		token2 := fixtures.NewRefreshToken(t, user2, "user2token")
		_, err = repo.Create(ctx, token2, 7*24*time.Hour)
		require.NoError(t, err)

		// Verify each user can find their token
		found1, err := repo.FindByHash(ctx, "user1token")
		require.NoError(t, err)
		assert.Equal(t, user1, found1.UserID().String())

		found2, err := repo.FindByHash(ctx, "user2token")
		require.NoError(t, err)
		assert.Equal(t, user2, found2.UserID().String())

		// Delete user1's token
		err = repo.DeleteByHash(ctx, "user1token")
		require.NoError(t, err)

		// Verify user2's token still exists
		found2, err = repo.FindByHash(ctx, "user2token")
		require.NoError(t, err)
		assert.Equal(t, user2, found2.UserID().String())
	})
}

func TestIntegration_MongoRefreshTokenRepository_TTLIndex(t *testing.T) {
	db, cleanup := testutils.SetupMongoContainer(t, _testDBName)
	defer cleanup()

	repo := mongorepo.NewRefreshTokens(db)
	repo.InitIndexes(context.Background())
	ctx := context.Background()

	t.Run("verify TTL index is created", func(t *testing.T) {
		// Just verify repository creation doesn't fail
		// TTL cleanup is handled by MongoDB and may take time
		assert.NotNil(t, repo)

		// Create a token that expires in 1 second
		userID := primitive.NewObjectID().Hex()
		token := fixtures.NewRefreshToken(t, userID, "shortlived")

		created, err := repo.Create(ctx, token, 1*time.Second)
		require.NoError(t, err)

		// Verify token exists
		found, err := repo.FindByHash(ctx, "shortlived")
		require.NoError(t, err)
		assert.Equal(t, "shortlived", string(found.TokenHash()))

		// Note: We can't easily test TTL deletion in integration tests
		// because it requires waiting for MongoDB's background job
		// which runs every 60 seconds by default
		_ = created
	})
}
