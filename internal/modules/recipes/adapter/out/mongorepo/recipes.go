package mongorepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/ports"
)

const collectionName = "recipes"

// Verify MongoRepository implements ports.RecipeRepository.
var _ ports.RecipeRepository = (*RecipeRepository)(nil)

type RecipeRepository struct {
	collection *mongo.Collection
}

func NewRecipes(db *mongo.Database) *RecipeRepository {
	return &RecipeRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *RecipeRepository) Create(
	ctx context.Context,
	recipe *entity.Recipe,
) (*entity.Recipe, error) {
	recipeModel, err := recipeModelFromDomain(recipe)
	if err != nil {
		return nil, err
	}
	recipeModel.prepareForInsert()

	_, err = r.collection.InsertOne(ctx, recipeModel)
	if err != nil {
		return nil, r.wrapError(err, "create recipe")
	}
	return recipeModel.toDomain()
}

func (r *RecipeRepository) FindByID(ctx context.Context, id string) (*entity.Recipe, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	var model recipeModel
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&model)
	if err != nil {
		return nil, r.wrapError(err, "find recipe by id")
	}

	return model.toDomain()
}

func (r *RecipeRepository) FindAll(ctx context.Context) ([]entity.Recipe, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, r.wrapError(err, "find all recipes")
	}
	defer cursor.Close(ctx)

	var recipeModels []recipeModel
	if err = cursor.All(ctx, &recipeModels); err != nil {
		return nil, r.wrapError(err, "decode recipes")
	}

	recipes := make([]entity.Recipe, len(recipeModels))
	for i, recipeModel := range recipeModels {
		var recipe *entity.Recipe
		recipe, err = recipeModel.toDomain()
		if err != nil {
			return nil, err
		}
		recipes[i] = *recipe
	}

	return recipes, nil
}

// Search searches recipes by title (case-insensitive).
func (r *RecipeRepository) Search(ctx context.Context, query string) ([]entity.Recipe, error) {
	filter := bson.M{
		"title": bson.M{
			"$regex":   query,
			"$options": "i", // case-insensitive
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, r.wrapError(err, "search recipes")
	}
	defer cursor.Close(ctx)

	var recipeModels []recipeModel
	if err = cursor.All(ctx, &recipeModels); err != nil {
		return nil, r.wrapError(err, "decode search results")
	}

	recipes := make([]entity.Recipe, len(recipeModels))
	for i, recipeModel := range recipeModels {
		var recipe *entity.Recipe
		recipe, err = recipeModel.toDomain()
		if err != nil {
			return nil, err
		}
		recipes[i] = *recipe
	}

	return recipes, nil
}

func (r *RecipeRepository) Update(
	ctx context.Context,
	recipe *entity.Recipe,
) (*entity.Recipe, error) {
	recipeModel, err := recipeModelFromDomain(recipe)
	if err != nil {
		return nil, err
	}
	recipeModel.prepareForUpdate()

	filter := bson.M{"_id": recipeModel.ID}
	update := bson.M{"$set": recipeModel}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, r.wrapError(err, "update recipe")
	}

	if result.MatchedCount == 0 {
		return nil, domain.ErrNotFound
	}

	return recipeModel.toDomain()
}

// Delete removes a recipe by its ID.
func (r *RecipeRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrNotFound
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return r.wrapError(err, "delete recipe")
	}

	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// wrapError converts MongoDB errors to domain errors.
func (r *RecipeRepository) wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// MongoDB specific errors
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.ErrNotFound
	}

	if mongo.IsTimeout(err) {
		return fmt.Errorf("%w: %s: %w", domain.ErrTimeout, operation, err)
	}

	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%w: %s: %w", domain.ErrConflict, operation, err)
	}

	return fmt.Errorf("%w: %s: %w", domain.ErrDatabase, operation, err)
}
