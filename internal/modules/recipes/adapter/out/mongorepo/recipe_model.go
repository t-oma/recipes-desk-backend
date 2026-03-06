package mongorepo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

type recipeModel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` //nolint:tagliatelle // mongoDB id
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Ingredients []ingredientModel  `bson:"ingredients"`
	Steps       []stepModel        `bson:"steps"`
	CookingTime int64              `bson:"cookingTime"`
	Portions    int                `bson:"portions"`
	Tags        []tagModel         `bson:"tags"`
	AuthorID    string             `bson:"authorId"`
	CreatedAt   time.Time          `bson:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"`
}

func (m *recipeModel) setTimestamps() {
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}

func (m *recipeModel) setID() {
	m.ID = primitive.NewObjectID()
}

func (m *recipeModel) prepareForInsert() {
	m.setTimestamps()
	if m.ID.IsZero() {
		m.setID()
	}
}

func (m *recipeModel) prepareForUpdate() {
	m.setTimestamps()
}

func (m *recipeModel) toDomain() (*entity.Recipe, error) {
	ingredients := make([]valueobject.Ingredient, len(m.Ingredients))
	for i, ingredient := range m.Ingredients {
		ingredients[i] = ingredient.toDomain()
	}

	steps := make([]valueobject.Step, len(m.Steps))
	for i, step := range m.Steps {
		steps[i] = step.toDomain()
	}

	tags := make([]valueobject.Tag, len(m.Tags))
	for i, tag := range m.Tags {
		tags[i] = tag.toDomain()
	}

	recipe, err := entity.NewRecipe(
		m.ID.Hex(),
		m.Title,
		m.Description,
		ingredients,
		steps,
		m.CookingTime,
		m.Portions,
		tags,
		m.AuthorID,
	)
	if err != nil {
		return nil, err
	}
	recipe.RestoreFromPersistence(m.UpdatedAt, m.CreatedAt)
	return recipe, nil
}

func recipeModelFromDomain(recipe *entity.Recipe) *recipeModel {
	var id primitive.ObjectID
	if recipe.HasID() {
		var err error
		id, err = primitive.ObjectIDFromHex(recipe.ID().String())
		if err != nil {
			id = primitive.NewObjectID()
		}
	}

	recipeIngredients := recipe.Ingredients()
	ingredients := make([]ingredientModel, len(recipeIngredients))
	for i, ingredient := range recipeIngredients {
		ingredients[i] = ingredientModelFromDomain(ingredient)
	}

	recipeSteps := recipe.Steps()
	steps := make([]stepModel, len(recipeSteps))
	for i, step := range recipeSteps {
		steps[i] = stepModelFromDomain(step)
	}

	recipeTags := recipe.Tags()
	tags := make([]tagModel, len(recipeTags))
	for i, tag := range recipeTags {
		tags[i] = tagModelFromDomain(tag)
	}

	return &recipeModel{
		ID:          id,
		Title:       recipe.Title().String(),
		Description: recipe.Description().String(),
		Ingredients: ingredients,
		Steps:       steps,
		CookingTime: recipe.CookingTime().SecondsInt64(),
		Portions:    recipe.Portions().Value(),
		Tags:        tags,
		AuthorID:    recipe.AuthorID().String(),
		CreatedAt:   recipe.CreatedAt(),
		UpdatedAt:   recipe.UpdatedAt(),
	}
}
