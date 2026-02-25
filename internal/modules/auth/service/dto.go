package service

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/auth/domain"
)

type RegisterParams struct {
	Email     string `json:"email"     binding:"required,email"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName"  binding:"required"`
	Password  string `json:"password"  binding:"required"`
}

type LoginParams struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SafeUser struct {
	ID                primitive.ObjectID `json:"id"`
	Email             string             `json:"email"`
	FirstName         string             `json:"firstName"`
	LastName          string             `json:"lastName"`
	CreatedAt         time.Time          `json:"createdAt"`
	PasswordUpdatedAt time.Time          `json:"passwordUpdatedAt"`
}

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

type AuthResult struct {
	User             *SafeUser `json:"user"`
	AccessToken      string    `json:"accessToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshToken     string    `json:"-"` // Not returned in JSON, only in cookie
	RefreshExpiresAt time.Time `json:"-"` // Not returned in JSON
}

// TokenResult contains the generated token and its metadata.
type TokenResult struct {
	Token     string
	ExpiresAt time.Time
}

type RefreshTokensParams struct {
	RefreshToken string
}

type RefreshTokensResult struct {
	AccessToken      string    `json:"accessToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshToken     string    `json:"refreshToken"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}
