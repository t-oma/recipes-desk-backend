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
)

const refreshCollectionName = "refresh_tokens"

// RefreshTokensRepository implements RefreshTokensRepository using MongoDB.
type RefreshTokensRepository struct {
	collection *mongo.Collection
}

var _ domain.RefreshTokensRepository = (*RefreshTokensRepository)(nil)

// NewRefreshTokens creates a new MongoRefreshTokenRepository.
func NewRefreshTokens(db *mongo.Database) *RefreshTokensRepository {
	return &RefreshTokensRepository{
		collection: db.Collection(refreshCollectionName),
	}
}

func (r *RefreshTokensRepository) InitIndexes(ctx context.Context) error {
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
func (r *RefreshTokensRepository) Create(
	ctx context.Context,
	token *domain.RefreshToken,
	ttl time.Duration,
) (*domain.RefreshToken, error) {
	// Convert to MongoDB model
	model, err := refreshTokenModelFromDomain(token)
	if err != nil {
		return nil, err
	}
	model.prepareForInsert(ttl)

	_, err = r.collection.InsertOne(ctx, model)
	if err != nil {
		return nil, err
	}

	// Update the domain model with the generated ID
	return model.toDomain(), nil
}

// FindByHash finds a refresh token by its hash.
func (r *RefreshTokensRepository) FindByHash(
	ctx context.Context,
	tokenHash string,
) (*domain.RefreshToken, error) {
	var model refreshTokenModel
	err := r.collection.FindOne(ctx, bson.M{"tokenHash": tokenHash}).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

// DeleteByHash deletes a refresh token by its hash.
func (r *RefreshTokensRepository) DeleteByHash(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"tokenHash": tokenHash})
	return err
}
