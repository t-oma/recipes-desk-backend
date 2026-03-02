package mongorepo

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
var _ domain.RecipesRepository = (*RecipesRepository)(nil)

type RecipesRepository struct {
	collection *mongo.Collection
}

func NewRecipes(db *mongo.Database) *RecipesRepository {
	return &RecipesRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *RecipesRepository) Create(
	ctx context.Context,
	recipe *domain.Recipe,
) (*domain.Recipe, error) {
	recipeModel := recipeModelFromDomain(recipe)
	recipeModel.prepareForInsert()

	_, err := r.collection.InsertOne(ctx, recipeModel)
	return recipeModel.toDomain(), err
}

func (r *RecipesRepository) FindByID(ctx context.Context, id string) (*domain.Recipe, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	var model recipeModel
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return model.toDomain(), nil
}

func (r *RecipesRepository) FindAll(ctx context.Context) ([]domain.Recipe, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var recipeModels []recipeModel
	if err = cursor.All(ctx, &recipeModels); err != nil {
		return nil, err
	}

	recipes := make([]domain.Recipe, len(recipeModels))
	for i, recipeModel := range recipeModels {
		recipes[i] = *recipeModel.toDomain()
	}

	return recipes, nil
}

// Search searches recipes by title (case-insensitive).
func (r *RecipesRepository) Search(ctx context.Context, query string) ([]domain.Recipe, error) {
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

	var recipeModels []recipeModel
	if err = cursor.All(ctx, &recipeModels); err != nil {
		return nil, err
	}

	recipes := make([]domain.Recipe, len(recipeModels))
	for i, recipeModel := range recipeModels {
		recipes[i] = *recipeModel.toDomain()
	}

	return recipes, nil
}

func (r *RecipesRepository) Update(
	ctx context.Context,
	recipe *domain.Recipe,
) (*domain.Recipe, error) {
	recipeModel := recipeModelFromDomain(recipe)
	recipeModel.prepareForUpdate()

	filter := bson.M{"_id": recipeModel.ID}
	update := bson.M{"$set": recipeModel}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, domain.ErrNotFound
	}

	return recipeModel.toDomain(), nil
}

// Delete removes a recipe by its ID.
func (r *RecipesRepository) Delete(ctx context.Context, id string) error {
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
