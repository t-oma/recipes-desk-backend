package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/auth/domain"
)

func TestRefreshToken_SetTimestamps(t *testing.T) {
	t.Run("sets timestamps on new token", func(t *testing.T) {
		token := domain.RefreshToken{ //nolint:exhaustruct // test struct
			UserID:    primitive.NewObjectID(),
			TokenHash: "somehash",
		}

		expiry := 7 * 24 * time.Hour
		before := time.Now()
		token.SetTimestamps(expiry)
		after := time.Now()

		assert.False(t, token.CreatedAt.IsZero())
		assert.False(t, token.ExpiresAt.IsZero())
		assert.True(t, token.CreatedAt.After(before) || token.CreatedAt.Equal(before))
		assert.True(t, token.CreatedAt.Before(after) || token.CreatedAt.Equal(after))

		// Check expiry is approximately correct
		expectedExpiry := token.CreatedAt.Add(expiry)
		assert.WithinDuration(t, expectedExpiry, token.ExpiresAt, time.Second)
	})

	t.Run("preserves existing CreatedAt", func(t *testing.T) {
		existingTime := time.Now().Add(-24 * time.Hour)
		token := domain.RefreshToken{
			ID:        primitive.NewObjectID(),
			UserID:    primitive.NewObjectID(),
			TokenHash: "somehash",
			CreatedAt: existingTime,
			ExpiresAt: time.Time{},
		}

		expiry := 7 * 24 * time.Hour
		token.SetTimestamps(expiry)

		// CreatedAt should be preserved
		assert.Equal(t, existingTime, token.CreatedAt)
		// ExpiresAt should be set based on existing CreatedAt (7 days later)
		assert.Equal(t, existingTime, token.CreatedAt)
		assert.True(t, token.ExpiresAt.After(token.CreatedAt))
		// ExpiresAt is always calculated from "now", not from CreatedAt
		assert.True(t, token.ExpiresAt.After(time.Now().Add(expiry-time.Hour)))
		assert.True(t, token.ExpiresAt.Before(time.Now().Add(expiry+time.Hour)))
	})

	t.Run("updates ExpiresAt even if already set", func(t *testing.T) {
		existingExpiry := time.Now().Add(30 * 24 * time.Hour)
		token := domain.RefreshToken{ //nolint:exhaustruct // test struct
			UserID:    primitive.NewObjectID(),
			TokenHash: "somehash",
			CreatedAt: time.Now(),
			ExpiresAt: existingExpiry,
		}

		newExpiry := 7 * 24 * time.Hour
		token.SetTimestamps(newExpiry)

		// ExpiresAt should be recalculated from now
		assert.NotEqual(t, existingExpiry, token.ExpiresAt)
		expectedExpiry := token.CreatedAt.Add(newExpiry)
		assert.WithinDuration(t, expectedExpiry, token.ExpiresAt, time.Second)
	})
}
