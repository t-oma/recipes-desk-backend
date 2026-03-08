package mongorepo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/ports"
)

const usersCollectionName = "users"

// UserRepository implements domain.Repository using MongoDB.
type UserRepository struct {
	collection *mongo.Collection
}

var _ ports.UserRepository = (*UserRepository)(nil)

// NewUsers creates a new MongoRepository.
func NewUsers(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection(usersCollectionName),
	}
}

// Create stores a new user in the database.
func (r *UserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	model := userModelFromDomain(user)
	model.prepareForInsert()

	_, err := r.collection.InsertOne(ctx, model)
	if err != nil {
		return nil, err
	}
	return model.toDomain()
}

// FindByID finds a user by their ID.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	filter := bson.M{"_id": objectID}

	var model userModel
	err = r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return model.toDomain()
}

// FindByEmail finds a user by their email.
func (r *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}

	var model userModel
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return model.toDomain()
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}

	num, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return num > 0, nil
}
