package in

import (
	"context"

	"recipes-desk/internal/modules/auth/application/dto"
)

// AuthService defines the interface for auth business logic.
type AuthService interface {
	GetByID(ctx context.Context, id string) (*dto.User, error)
	Register(ctx context.Context, params dto.RegisterInput) (*dto.AuthResult, error)
	Login(ctx context.Context, params dto.LoginInput) (*dto.AuthResult, error)
	RefreshTokens(
		ctx context.Context,
		input dto.RefreshTokensInput,
	) (*dto.RefreshTokensResult, error)
}
