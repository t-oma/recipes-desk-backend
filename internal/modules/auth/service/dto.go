package service

import (
	"time"

	"recipes-desk/internal/modules/auth/domain"
)

// RegisterInput contains parameters for user registration.
type RegisterInput struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

// LoginInput contains parameters for user login.
type LoginInput struct {
	Email    string
	Password string
}

// UserDTO represents a user without sensitive information.
type UserDTO struct {
	ID                string
	Email             string
	FirstName         string
	LastName          string
	CreatedAt         time.Time
	PasswordUpdatedAt time.Time
}

// toUserDTO converts a domain user to a safe user.
func toUserDTO(user *domain.User) *UserDTO {
	return &UserDTO{
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
	User             *UserDTO
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

// RefreshTokensInput contains parameters for token refresh.
type RefreshTokensInput struct {
	RefreshToken string
}

// RefreshTokensResult contains the result of a successful token refresh.
type RefreshTokensResult struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}
