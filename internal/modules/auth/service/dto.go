package service

import (
	"time"

	"recipes-desk/internal/modules/auth/domain"
)

// RegisterParams contains parameters for user registration.
type RegisterParams struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

// LoginParams contains parameters for user login.
type LoginParams struct {
	Email    string
	Password string
}

// SafeUser represents a user without sensitive information.
type SafeUser struct {
	ID                string
	Email             string
	FirstName         string
	LastName          string
	CreatedAt         time.Time
	PasswordUpdatedAt time.Time
}

// SafeUserFromUser converts a domain user to a safe user.
func SafeUserFromUser(user *domain.User) *SafeUser {
	return &SafeUser{
		ID:                user.ID,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		CreatedAt:         user.CreatedAt,
		PasswordUpdatedAt: user.PasswordUpdatedAt,
	}
}

// AuthResult contains the result of a successful authentication.
type AuthResult struct {
	User             *SafeUser
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

// RefreshTokensParams contains parameters for token refresh.
type RefreshTokensParams struct {
	RefreshToken string
}

// RefreshTokensResult contains the result of a successful token refresh.
type RefreshTokensResult struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}
