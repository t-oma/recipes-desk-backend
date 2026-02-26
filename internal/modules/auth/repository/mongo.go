package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/auth/domain"
)

const collectionName = "users"

type MongoRepository struct {
	collection *mongo.Collection
}

var _ domain.Repository = (*MongoRepository)(nil)

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *MongoRepository) Create(ctx context.Context, user *domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	user.SetTimestamps()

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *MongoRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	filter := bson.M{"_id": objectID}

	var user domain.User
	err = r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *MongoRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}

	var user domain.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *MongoRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}

	num, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return num > 0, nil
}
