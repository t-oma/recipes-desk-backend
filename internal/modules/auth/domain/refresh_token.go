package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"-"` //nolint:tagliatelle // MongoDB _id field
	UserID    primitive.ObjectID `bson:"userId"        json:"-"`
	TokenHash string             `bson:"tokenHash"     json:"-"`
	ExpiresAt time.Time          `bson:"expiresAt"     json:"-"`
	CreatedAt time.Time          `bson:"createdAt"     json:"-"`
}

// SetTimestamps sets the creation and expiration timestamps.
func (rt *RefreshToken) SetTimestamps(expiry time.Duration) {
	now := time.Now()
	if rt.CreatedAt.IsZero() { // Set only if not set
		rt.CreatedAt = now
	}
	rt.ExpiresAt = now.Add(expiry)
}
