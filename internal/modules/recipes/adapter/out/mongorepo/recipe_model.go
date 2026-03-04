package mongorepo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/recipes/domain"
)

type recipeModel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` //nolint:tagliatelle // mongoDB id
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Ingredients []ingredientModel  `bson:"ingredients"`
	Steps       []stepModel        `bson:"steps"`
	CookingTime int                `bson:"cookingTime"`
	Portions    int                `bson:"portions"`
	Tags        []string           `bson:"tags"`
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
	m.setID()
}

func (m *recipeModel) prepareForUpdate() {
	m.setTimestamps()
}

func (m *recipeModel) toDomain() *domain.Recipe {
	ingredients := make([]domain.Ingredient, len(m.Ingredients))
	for i, ingredient := range m.Ingredients {
		ingredients[i] = ingredient.toDomain()
	}

	steps := make([]domain.Step, len(m.Steps))
	for i, step := range m.Steps {
		steps[i] = step.toDomain()
	}

	return &domain.Recipe{
		ID:          m.ID.Hex(),
		Title:       m.Title,
		Description: m.Description,
		Ingredients: ingredients,
		Steps:       steps,
		CookingTime: m.CookingTime,
		Portions:    m.Portions,
		Tags:        m.Tags,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func recipeModelFromDomain(recipe *domain.Recipe) *recipeModel {
	var id primitive.ObjectID
	if recipe.ID != "" {
		var err error
		id, err = primitive.ObjectIDFromHex(recipe.ID)
		if err != nil {
			id = primitive.NilObjectID
		}
	}

	ingredients := make([]ingredientModel, len(recipe.Ingredients))
	for i, ingredient := range recipe.Ingredients {
		ingredients[i] = ingredientModelFromDomain(ingredient)
	}

	steps := make([]stepModel, len(recipe.Steps))
	for i, step := range recipe.Steps {
		steps[i] = stepModelFromDomain(step)
	}

	return &recipeModel{
		ID:          id,
		Title:       recipe.Title,
		Description: recipe.Description,
		Ingredients: ingredients,
		Steps:       steps,
		CookingTime: recipe.CookingTime,
		Portions:    recipe.Portions,
		Tags:        recipe.Tags,
		CreatedAt:   recipe.CreatedAt,
		UpdatedAt:   recipe.UpdatedAt,
	}
}
