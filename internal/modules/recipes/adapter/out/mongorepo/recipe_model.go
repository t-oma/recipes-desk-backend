package mongorepo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
	"recipes-desk/pkg/sliceutils"
)

type recipeModel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` //nolint:tagliatelle // mongoDB id
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Ingredients []ingredientModel  `bson:"ingredients"`
	Steps       []stepModel        `bson:"steps"`
	CookingTime int64              `bson:"cookingTime"`
	Portions    int                `bson:"portions"`
	Tags        []tagNameModel     `bson:"tags"`
	AuthorID    primitive.ObjectID `bson:"authorId"`
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
	var err error
	id, err := valueobject.NewRecipeID(m.ID.Hex())
	if err != nil {
		return nil, err
	}
	title, err := valueobject.NewTitle(m.Title)
	if err != nil {
		return nil, err
	}
	desc, err := valueobject.NewDescription(m.Description)
	if err != nil {
		return nil, err
	}
	cookingTime, err := valueobject.NewCookingTime(m.CookingTime)
	if err != nil {
		return nil, err
	}
	portions, err := valueobject.NewPortions(m.Portions)
	if err != nil {
		return nil, err
	}
	authorID, err := valueobject.NewAuthorID(m.AuthorID.Hex())
	if err != nil {
		return nil, err
	}

	ingredients, err := sliceutils.MapSliceWithErr(
		m.Ingredients,
		func(ing ingredientModel) (valueobject.Ingredient, error) {
			return ing.toDomain()
		},
	)
	if err != nil {
		return nil, err
	}

	steps, err := sliceutils.MapSliceWithErr(
		m.Steps,
		func(step stepModel) (valueobject.Step, error) {
			return step.toDomain()
		},
	)
	if err != nil {
		return nil, err
	}

	tags, err := sliceutils.MapSliceWithErr(
		m.Tags,
		func(tag tagNameModel) (valueobject.TagName, error) {
			return tag.toDomain()
		},
	)
	if err != nil {
		return nil, err
	}

	recipe, err := entity.NewRecipe(
		id,
		title,
		desc,
		ingredients,
		steps,
		cookingTime,
		portions,
		tags,
		authorID,
	)
	if err != nil {
		return nil, err
	}
	recipe.RestoreFromPersistence(m.UpdatedAt, m.CreatedAt)
	return recipe, nil
}

func recipeModelFromDomain(recipe *entity.Recipe) (*recipeModel, error) {
	var id primitive.ObjectID
	if recipe.HasID() {
		var err error
		id, err = primitive.ObjectIDFromHex(recipe.ID().String())
		if err != nil {
			id = primitive.NewObjectID()
		}
	}
	authorID, err := primitive.ObjectIDFromHex(recipe.AuthorID().String())
	if err != nil {
		return nil, err
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
	tags := make([]tagNameModel, len(recipeTags))
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
		AuthorID:    authorID,
		CreatedAt:   recipe.CreatedAt(),
		UpdatedAt:   recipe.UpdatedAt(),
	}, nil
}
