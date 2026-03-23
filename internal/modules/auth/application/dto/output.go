package dto

import (
	"time"
)

// User represents a user without sensitive information.
type User struct {
	ID                string
	Email             string
	FirstName         string
	LastName          string
	CreatedAt         time.Time
	PasswordUpdatedAt time.Time
}

// AuthResult contains the result of a successful authentication.
type AuthResult struct {
	User             *User
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// TokenResult contains the generated token and its metadata.
type TokenResult struct {
	Token     string
	ExpiresAt time.Time
}

// RefreshTokensResult contains the result of a successful token refresh.
type RefreshTokensResult struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}
