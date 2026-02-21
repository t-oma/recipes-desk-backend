package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/recipes/domain"
)

const collectionName = "recipes"

// Verify MongoRepository implements domain.Repository.
var _ domain.Repository = (*MongoRepository)(nil)

// MongoRepository implements domain.Repository using MongoDB.
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection(collectionName),
	}
}

// Create inserts a new recipe into the database.
func (r *MongoRepository) Create(ctx context.Context, recipe *domain.Recipe) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if recipe.ID.IsZero() {
		recipe.ID = primitive.NewObjectID()
	}
	recipe.SetTimestamps()

	_, err := r.collection.InsertOne(ctx, recipe)
	return err
}

// FindByID finds a recipe by its ID.
func (r *MongoRepository) FindByID(ctx context.Context, id string) (*domain.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	var recipe domain.Recipe
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&recipe)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &recipe, nil
}

// FindAll returns all recipes.
func (r *MongoRepository) FindAll(ctx context.Context) ([]domain.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var recipes []domain.Recipe
	if err = cursor.All(ctx, &recipes); err != nil {
		return nil, err
	}

	return recipes, nil
}

// Search searches recipes by title (case-insensitive).
func (r *MongoRepository) Search(ctx context.Context, query string) ([]domain.Recipe, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{
		"title": bson.M{
			"$regex":   query,
			"$options": "i", // case-insensitive
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var recipes []domain.Recipe
	if err = cursor.All(ctx, &recipes); err != nil {
		return nil, err
	}

	return recipes, nil
}

// Update updates an existing recipe.
func (r *MongoRepository) Update(ctx context.Context, recipe *domain.Recipe) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	recipe.SetTimestamps()

	filter := bson.M{"_id": recipe.ID}
	update := bson.M{"$set": recipe}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete removes a recipe by its ID.
func (r *MongoRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrNotFound
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}
