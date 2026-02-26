package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"recipes-desk/internal/modules/auth/domain"
)

// MongoRefreshTokenRepository implements RefreshTokenRepository using MongoDB.
type MongoRefreshTokenRepository struct {
	collection *mongo.Collection
}

var _ domain.RefreshTokenRepository = (*MongoRefreshTokenRepository)(nil)

// NewMongoRefreshTokenRepository creates a new MongoRefreshTokenRepository.
func NewMongoRefreshTokenRepository(db *mongo.Database) *MongoRefreshTokenRepository {
	repo := &MongoRefreshTokenRepository{
		collection: db.Collection("refresh_tokens"),
	}

	// Create TTL index on expiresAt field
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	_, err := repo.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		fmt.Printf("Failed to create TTL index: %v\n", err)
	}

	return repo
}

// Create stores a new refresh token in the database.
func (r *MongoRefreshTokenRepository) Create(
	ctx context.Context,
	token *domain.RefreshToken,
) error {
	if token.ID.IsZero() {
		token.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return err
	}
	return nil
}

// FindByHash finds a refresh token by its hash.
func (r *MongoRefreshTokenRepository) FindByHash(
	ctx context.Context,
	tokenHash string,
) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{"tokenHash": tokenHash}).Decode(&token)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &token, nil
}

// DeleteByHash deletes a refresh token by its hash.
func (r *MongoRefreshTokenRepository) DeleteByHash(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"tokenHash": tokenHash})
	return err
}
