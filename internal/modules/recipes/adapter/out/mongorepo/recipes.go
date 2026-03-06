package mongorepo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

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
	recipeModel := recipeModelFromDomain(recipe)
	recipeModel.prepareForInsert()

	_, err := r.collection.InsertOne(ctx, recipeModel)
	if err != nil {
		return nil, err
	}
	return recipeModel.toDomain()
}

func (r *RecipeRepository) FindByID(ctx context.Context, id string) (*entity.Recipe, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ports.ErrNotFound
	}

	var model recipeModel
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&model)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	return model.toDomain()
}

func (r *RecipeRepository) FindAll(ctx context.Context) ([]entity.Recipe, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var recipeModels []recipeModel
	if err = cursor.All(ctx, &recipeModels); err != nil {
		return nil, err
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
		return nil, err
	}
	defer cursor.Close(ctx)

	var recipeModels []recipeModel
	if err = cursor.All(ctx, &recipeModels); err != nil {
		return nil, err
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
	recipeModel := recipeModelFromDomain(recipe)
	recipeModel.prepareForUpdate()

	filter := bson.M{"_id": recipeModel.ID}
	update := bson.M{"$set": recipeModel}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, ports.ErrNotFound
	}

	return recipeModel.toDomain()
}

// Delete removes a recipe by its ID.
func (r *RecipeRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ports.ErrNotFound
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ports.ErrNotFound
	}

	return nil
}
