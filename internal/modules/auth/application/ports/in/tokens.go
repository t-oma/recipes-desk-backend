package in

import (
	"context"

	"recipes-desk/internal/modules/auth/application/dto"
)

type TokenService interface {
	GenerateAccessToken(userID string) (*dto.TokenResult, error)
	ValidateAccessToken(tokenString string) (*dto.Claims, error)
	GenerateRefreshToken(ctx context.Context, userID string) (*dto.TokenResult, error)
	ValidateRefreshToken(ctx context.Context, plainToken string) (string, error)
	RotateRefreshToken(
		ctx context.Context,
		oldToken string,
	) (*dto.TokenResult, string, error)
}
