package ports

import "context"

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}
