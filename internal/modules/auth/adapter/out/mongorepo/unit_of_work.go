package mongorepo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/auth/domain/ports"
)

type UnitOfWork struct {
	client *mongo.Client
}

var _ ports.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(
	client *mongo.Client,
) *UnitOfWork {
	return &UnitOfWork{
		client: client,
	}
}

func (u *UnitOfWork) Execute(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	session, err := u.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	//nolint:contextcheck // sesCtx is mongo.SessionContext which wraps context.Context
	_, err = session.WithTransaction(ctx, func(sesCtx mongo.SessionContext) (any, error) {
		return nil, fn(sesCtx)
	})
	return err
}
