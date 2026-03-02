package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrTokenNotFound = errors.New("refresh token not found")
)

// RefreshTokensRepository defines the interface for refresh token storage.
type RefreshTokensRepository interface {
	Create(ctx context.Context, token *RefreshToken, ttl time.Duration) (*RefreshToken, error)
	FindByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	DeleteByHash(ctx context.Context, tokenHash string) error
}
