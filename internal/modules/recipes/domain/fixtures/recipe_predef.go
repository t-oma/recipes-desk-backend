package fixtures

import (
	"testing"

	"github.com/stretchr/testify/require"

	vo "recipes-desk/internal/modules/recipes/domain/valueobject"
)

func ValidIngredients(t *testing.T) []vo.Ingredient {
	t.Helper()

	ing1, err := vo.NewIngredient("Flour", 500, vo.UnitGram)
	require.NoError(t, err)
	ing2, err := vo.NewIngredient("Eggs", 3, vo.UnitPcs)
	require.NoError(t, err)
	return []vo.Ingredient{ing1, ing2}
}

func ValidSteps(t *testing.T) []vo.Step {
	t.Helper()

	step1, err := vo.NewStep(1, "Mix ingredients", 5)
	require.NoError(t, err)
	step2, err := vo.NewStep(2, "Bake", 30)
	require.NoError(t, err)
	return []vo.Step{step1, step2}
}

func ValidTags(t *testing.T) []vo.Tag {
	t.Helper()

	tag1, err := vo.NewTag("italian")
	require.NoError(t, err)
	tag2, err := vo.NewTag("pasta")
	require.NoError(t, err)
	return []vo.Tag{tag1, tag2}
}

func ValidID(t *testing.T) vo.EntityID {
	t.Helper()

	id, err := vo.NewEntityID("id123")
	require.NoError(t, err)
	return id
}

func ValidTitle(t *testing.T) vo.Title {
	t.Helper()

	title, err := vo.NewTitle("Pasta Carbonara")
	require.NoError(t, err)
	return title
}

func ValidDescription(t *testing.T) vo.Description {
	t.Helper()

	desc, err := vo.NewDescription("Classic Italian pasta dish with eggs and cheese")
	require.NoError(t, err)
	return desc
}

func ValidCookingTime(t *testing.T) vo.CookingTime {
	t.Helper()

	cookingTime, err := vo.NewCookingTime(30)
	require.NoError(t, err)
	return cookingTime
}

func ValidPortions(t *testing.T) vo.Portions {
	t.Helper()

	portions, err := vo.NewPortions(4)
	require.NoError(t, err)
	return portions
}

func ValidAuthorID(t *testing.T) vo.AuthorID {
	t.Helper()

	authorID, err := vo.NewAuthorID("author123")
	require.NoError(t, err)
	return authorID
}
