package fixtures

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

// NewRecipe creates a recipe with valid test data.
func NewRecipe(t *testing.T, id, authorID, title string) *entity.Recipe {
	t.Helper()

	ingredients := []valueobject.Ingredient{
		MustIngredient(t, "Flour", 500, "g"),
		MustIngredient(t, "Eggs", 3, "pcs"),
	}
	steps := []valueobject.Step{
		MustStep(t, 1, "Mix ingredients", 300),
		MustStep(t, 2, "Bake", 1800),
	}
	tags := []valueobject.Tag{
		MustTag(t, "test"),
		MustTag(t, "integration"),
	}

	idVO, err := valueobject.NewRecipeID(id)
	require.NoError(t, err)
	titleVO, err := valueobject.NewTitle(title)
	require.NoError(t, err)
	descVO, err := valueobject.NewDescription(
		"This is a valid description for integration test with at least 10 characters",
	)
	require.NoError(t, err)
	cookingTimeVO, err := valueobject.NewCookingTime(35 * 60)
	require.NoError(t, err)
	portionsVO, err := valueobject.NewPortions(4)
	require.NoError(t, err)
	authorIDVO, err := valueobject.NewAuthorID(authorID)
	require.NoError(t, err)

	entity, err := entity.NewRecipe(
		idVO,
		titleVO,
		descVO,
		ingredients,
		steps,
		cookingTimeVO,
		portionsVO,
		tags,
		authorIDVO,
	)
	require.NoError(t, err)

	return entity
}

// MustIngredient creates an ingredient or fails the test.
func MustIngredient(t *testing.T, name string, amount float64, unit string) valueobject.Ingredient {
	t.Helper()
	ing, err := valueobject.NewIngredient(name, amount, unit)
	require.NoError(t, err)
	return ing
}

// MustStep creates a step or fails the test.
func MustStep(t *testing.T, order int, description string, durationSec int64) valueobject.Step {
	t.Helper()
	step, err := valueobject.NewStep(order, description, durationSec)
	require.NoError(t, err)
	return step
}

// MustTag creates a tag or fails the test.
func MustTag(t *testing.T, name string) valueobject.Tag {
	t.Helper()
	tag, err := valueobject.NewTag(name)
	require.NoError(t, err)
	return tag
}
