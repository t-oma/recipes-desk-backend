package ports

import (
	"context"
	"time"

	"recipes-desk/internal/modules/auth/domain/entity"
)

type RefreshTokenRepository interface {
	Create(
		ctx context.Context,
		token *entity.RefreshToken,
		ttl time.Duration,
	) (*entity.RefreshToken, error)
	FindByHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	DeleteByHash(ctx context.Context, tokenHash string) error
}
