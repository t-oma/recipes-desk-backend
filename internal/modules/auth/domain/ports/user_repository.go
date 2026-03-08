package ports

import (
	"context"

	"recipes-desk/internal/modules/auth/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
