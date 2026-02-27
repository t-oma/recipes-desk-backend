package domain

import (
	"time"
)

// RefreshToken represents a refresh token in the domain layer.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}
