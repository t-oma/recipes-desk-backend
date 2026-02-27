package mongorepo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/auth/domain"
)

const usersCollectionName = "users"

// UsersRepository implements domain.Repository using MongoDB.
type UsersRepository struct {
	collection *mongo.Collection
}

var _ domain.UsersRepository = (*UsersRepository)(nil)

// NewUsers creates a new MongoRepository.
func NewUsers(db *mongo.Database) *UsersRepository {
	return &UsersRepository{
		collection: db.Collection(usersCollectionName),
	}
}

// Create stores a new user in the database.
func (r *UsersRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	model := userModelFromDomain(user)
	model.setTimestamps()
	model.setID()

	_, err := r.collection.InsertOne(ctx, model)
	return model.toDomain(), err
}

// FindByID finds a user by their ID.
func (r *UsersRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	filter := bson.M{"_id": objectID}

	var model userModel
	err = r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return model.toDomain(), nil
}

// FindByEmail finds a user by their email.
func (r *UsersRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}

	var model userModel
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return model.toDomain(), nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UsersRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}

	num, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return num > 0, nil
}
