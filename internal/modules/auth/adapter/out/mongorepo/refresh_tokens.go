package mongorepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/ports"
)

const refreshCollectionName = "refresh_tokens"

// RefreshTokenRepository implements RefreshTokenRepository using MongoDB.
type RefreshTokenRepository struct {
	collection *mongo.Collection
}

var _ ports.RefreshTokenRepository = (*RefreshTokenRepository)(nil)

// NewRefreshTokens creates a new MongoRefreshTokenRepository.
func NewRefreshTokens(db *mongo.Database) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		collection: db.Collection(refreshCollectionName),
	}
}

func (r *RefreshTokenRepository) InitIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		fmt.Printf("Failed to create TTL index: %v\n", err)
		return err
	}

	return nil
}

// Create stores a new refresh token in the database.
func (r *RefreshTokenRepository) Create(
	ctx context.Context,
	token *entity.RefreshToken,
	ttl time.Duration,
) (*entity.RefreshToken, error) {
	model, err := refreshTokenModelFromDomain(token)
	if err != nil {
		return nil, err
	}
	model.prepareForInsert(ttl)

	_, err = r.collection.InsertOne(ctx, model)
	if err != nil {
		return nil, err
	}

	return model.toDomain()
}

// FindByHash finds a refresh token by its hash.
func (r *RefreshTokenRepository) FindByHash(
	ctx context.Context,
	tokenHash string,
) (*entity.RefreshToken, error) {
	var model refreshTokenModel
	err := r.collection.FindOne(ctx, bson.M{"tokenHash": tokenHash}).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}
	return model.toDomain()
}

// DeleteByHash deletes a refresh token by its hash.
func (r *RefreshTokenRepository) DeleteByHash(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"tokenHash": tokenHash})
	return err
}
