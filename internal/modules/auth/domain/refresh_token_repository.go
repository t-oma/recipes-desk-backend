package domain

import (
	"context"
	"errors"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrTokenNotFound = errors.New("refresh token not found")
)

// RefreshTokenRepository defines the interface for refresh token storage.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	DeleteByHash(ctx context.Context, tokenHash string) error
}
