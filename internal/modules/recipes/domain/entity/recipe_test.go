package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/domain/entity"
	fxt "recipes-desk/internal/modules/recipes/domain/fixtures"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

func TestNewRecipe(t *testing.T) {
	tests := []struct {
		name            string
		makeIngredients func(t *testing.T) []valueobject.Ingredient
		makeSteps       func(t *testing.T) []valueobject.Step
		makeTags        func(t *testing.T) []valueobject.TagName
		wantErr         error
	}{
		{
			name:            "valid recipe",
			makeIngredients: fxt.ValidIngredients,
			makeSteps:       fxt.ValidSteps,
			makeTags:        fxt.ValidTags,
			wantErr:         nil,
		},
		{
			name: "no ingredients",
			makeIngredients: func(_ *testing.T) []valueobject.Ingredient {
				return []valueobject.Ingredient{}
			},
			makeSteps: fxt.ValidSteps,
			makeTags:  fxt.ValidTags,
			wantErr:   domain.ErrNoIngredients,
		},
		{
			name:            "no steps",
			makeIngredients: fxt.ValidIngredients,
			makeSteps: func(_ *testing.T) []valueobject.Step {
				return []valueobject.Step{}
			},
			makeTags: fxt.ValidTags,
			wantErr:  domain.ErrNoSteps,
		},
		{
			name:            "no tags",
			makeIngredients: fxt.ValidIngredients,
			makeSteps:       fxt.ValidSteps,
			makeTags: func(_ *testing.T) []valueobject.TagName {
				return []valueobject.TagName{}
			},
			wantErr: domain.ErrNoTags,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := entity.NewRecipe(
				fxt.ValidID(t),
				fxt.ValidTitle(t),
				fxt.ValidDescription(t),
				tt.makeIngredients(t),
				tt.makeSteps(t),
				fxt.ValidCookingTime(t),
				fxt.ValidPortions(t),
				tt.makeTags(t),
				fxt.ValidAuthorID(t),
			)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, recipe)

				require.Len(t, recipe.Ingredients(), len(tt.makeIngredients(t)))
				require.Len(t, recipe.Steps(), len(tt.makeSteps(t)))
				require.Len(t, recipe.Tags(), len(tt.makeTags(t)))
				require.True(t, recipe.HasID())
			}
		})
	}
}

func TestEntity_Equals(t *testing.T) {
	tests := []struct {
		name        string
		makeRecipes func(t *testing.T) (*entity.Recipe, *entity.Recipe)
		wantEqual   bool
	}{
		{
			name: "equal entities",
			makeRecipes: func(t *testing.T) (*entity.Recipe, *entity.Recipe) {
				t.Helper()

				recipe1, _ := entity.NewRecipe(
					fxt.ValidID(t),
					fxt.ValidTitle(t),
					fxt.ValidDescription(t),
					fxt.ValidIngredients(t),
					fxt.ValidSteps(t),
					fxt.ValidCookingTime(t),
					fxt.ValidPortions(t),
					fxt.ValidTags(t),
					fxt.ValidAuthorID(t),
				)
				recipe2, _ := entity.NewRecipe(
					fxt.ValidID(t),
					fxt.ValidTitle(t),
					fxt.ValidDescription(t),
					fxt.ValidIngredients(t),
					fxt.ValidSteps(t),
					fxt.ValidCookingTime(t),
					fxt.ValidPortions(t),
					fxt.ValidTags(t),
					fxt.ValidAuthorID(t),
				)
				return recipe1, recipe2
			},
			wantEqual: true,
		},
		{
			name: "not equal entities",
			makeRecipes: func(t *testing.T) (*entity.Recipe, *entity.Recipe) {
				t.Helper()

				recipe1, _ := entity.NewRecipe(
					fxt.ValidID(t),
					fxt.ValidTitle(t),
					fxt.ValidDescription(t),
					fxt.ValidIngredients(t),
					fxt.ValidSteps(t),
					fxt.ValidCookingTime(t),
					fxt.ValidPortions(t),
					fxt.ValidTags(t),
					fxt.ValidAuthorID(t),
				)
				diffIngredients := []valueobject.Ingredient{
					fxt.MustIngredient(t, "Milk", 100, "ml"),
				}
				diffID, err := valueobject.NewRecipeID("id456")
				require.NoError(t, err)

				recipe2, _ := entity.NewRecipe(
					diffID,
					fxt.ValidTitle(t),
					fxt.ValidDescription(t),
					diffIngredients,
					fxt.ValidSteps(t),
					fxt.ValidCookingTime(t),
					fxt.ValidPortions(t),
					fxt.ValidTags(t),
					fxt.ValidAuthorID(t),
				)
				return recipe1, recipe2
			},
			wantEqual: false,
		},
		{
			name: "not equal entities - nil",
			makeRecipes: func(t *testing.T) (*entity.Recipe, *entity.Recipe) {
				t.Helper()

				recipe1, err := entity.NewRecipe(
					fxt.ValidID(t),
					fxt.ValidTitle(t),
					fxt.ValidDescription(t),
					fxt.ValidIngredients(t),
					fxt.ValidSteps(t),
					fxt.ValidCookingTime(t),
					fxt.ValidPortions(t),
					fxt.ValidTags(t),
					fxt.ValidAuthorID(t),
				)
				require.NoError(t, err)
				return recipe1, nil
			},
			wantEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe1, recipe2 := tt.makeRecipes(t)

			if tt.wantEqual {
				require.True(t, recipe1.Equals(recipe2))
			} else {
				require.False(t, recipe1.Equals(recipe2))
			}
		})
	}
}

func TestRecipe_UpdateTitle(t *testing.T) {
	recipe := fxt.NewRecipe(t, "id123", "author123", "Test Recipe")

	t.Run("new title", func(t *testing.T) {
		newTitle, _ := valueobject.NewTitle("")

		recipe.UpdateTitle(newTitle)
		require.Equal(t, newTitle.String(), recipe.Title().String())
	})
}

func TestRecipe_UpdateDescription(t *testing.T) {
	recipe := fxt.NewRecipe(t, "id123", "author123", "Test Recipe")

	t.Run("new description", func(t *testing.T) {
		newDescription, _ := valueobject.NewDescription("")

		recipe.UpdateDescription(newDescription)
		require.Equal(t, newDescription.String(), recipe.Description().String())
	})
}

func TestRecipe_AddIngredient(t *testing.T) {
	recipe := fxt.NewRecipe(t, "id123", "author123", "Test Recipe")
	ingredients := recipe.Ingredients()

	t.Run("existing ingredient", func(t *testing.T) {
		ingredient, _ := valueobject.NewIngredient(ingredients[0].Name(), 500, "g")

		err := recipe.AddIngredient(ingredient)
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrDuplicateIngredient)
		require.Len(t, recipe.Ingredients(), len(ingredients))
	})

	t.Run("new ingredient", func(t *testing.T) {
		ingredient, _ := valueobject.NewIngredient("Milk", 500, "ml")

		recipe.AddIngredient(ingredient)
		require.Len(t, recipe.Ingredients(), len(ingredients)+1)
	})
}

func TestRecipe_AddStep(t *testing.T) {
	recipe := fxt.NewRecipe(t, "id123", "author123", "Test Recipe")
	steps := recipe.Steps()

	t.Run("wrong order", func(t *testing.T) {
		step, _ := valueobject.NewStep(0, "Mix ingredients", 60)

		err := recipe.AddStep(step)
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrValidation)
		require.Len(t, recipe.Steps(), len(steps))
	})

	t.Run("new step", func(t *testing.T) {
		step, _ := valueobject.NewStep(len(steps)+1, "Mix ingredients", 5)
		recipe.AddStep(step)

		require.Len(t, recipe.Steps(), len(steps)+1)
	})
}
