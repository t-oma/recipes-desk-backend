package mongorepo

import (
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/auth/domain"
)

func wrapMongoError(err error, operation string) error {
	if err == nil {
		return nil
	}

	if mongo.IsTimeout(err) {
		return fmt.Errorf("%w: %s: %w", domain.ErrTimeout, operation, err)
	}

	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%w: %s: %w", domain.ErrConflict, operation, err)
	}

	return fmt.Errorf("%w: %s: %w", domain.ErrDatabase, operation, err)
}
