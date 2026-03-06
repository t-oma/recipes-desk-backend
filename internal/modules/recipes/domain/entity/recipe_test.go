package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/fixtures"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
)

func TestNewEntity(t *testing.T) {
	ing1, _ := valueobject.NewIngredient("Flour", 500, valueobject.UnitGram)
	ing2, _ := valueobject.NewIngredient("Eggs", 3, valueobject.UnitPcs)
	validIngredients := []valueobject.Ingredient{ing1, ing2}

	step1, _ := valueobject.NewStep(1, "Mix ingredients", 5)
	step2, _ := valueobject.NewStep(2, "Bake", 30)
	validSteps := []valueobject.Step{step1, step2}

	tag1, _ := valueobject.NewTag("italian")
	tag2, _ := valueobject.NewTag("pasta")
	validTags := []valueobject.Tag{tag1, tag2}

	tests := []struct {
		name        string
		id          string
		title       string
		desc        string
		ingredients []valueobject.Ingredient
		steps       []valueobject.Step
		cookingTime int64
		portions    int
		tags        []valueobject.Tag
		authorID    string
		wantErr     error
	}{
		{
			name:        "valid recipe",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       validSteps,
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     nil,
		},
		{
			name:        "no ingredients",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: []valueobject.Ingredient{},
			steps:       validSteps,
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     entity.ErrNoIngredients,
		},
		{
			name:        "no steps",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       []valueobject.Step{},
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     entity.ErrNoSteps,
		},
		{
			name:        "no tags",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       validSteps,
			cookingTime: 30,
			portions:    4,
			tags:        []valueobject.Tag{},
			authorID:    "author123",
			wantErr:     entity.ErrNoTags,
		},
		{
			name:        "empty id",
			id:          "",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       validSteps,
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     valueobject.ErrRecipeIDEmpty,
		},
		{
			name:        "empty title",
			id:          "id123",
			title:       "",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       validSteps,
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     valueobject.ErrTitleEmpty,
		},
		{
			name:        "empty description",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "",
			ingredients: validIngredients,
			steps:       validSteps,
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     valueobject.ErrDescriptionEmpty,
		},
		{
			name:        "invalid cooking time",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       validSteps,
			cookingTime: -1,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     valueobject.ErrCookingNegativeTime,
		},
		{
			name:        "invalid portions",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       validSteps,
			cookingTime: 30,
			portions:    -1,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     valueobject.ErrPortionsTooFew,
		},
		{
			name:        "empty author ID",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       validSteps,
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "",
			wantErr:     valueobject.ErrAuthorIDEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := entity.NewRecipe(
				tt.id,
				tt.title,
				tt.desc,
				tt.ingredients,
				tt.steps,
				tt.cookingTime,
				tt.portions,
				tt.tags,
				tt.authorID,
			)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, recipe)

				require.Equal(t, tt.title, recipe.Title().String())
				require.Equal(t, tt.desc, recipe.Description().String())
				require.Len(t, recipe.Ingredients(), len(tt.ingredients))
				require.Len(t, recipe.Steps(), len(tt.steps))
				require.Equal(t, tt.cookingTime, recipe.CookingTime().SecondsInt64())
				require.Equal(t, tt.portions, recipe.Portions().Value())
				require.Len(t, recipe.Tags(), len(tt.tags))
				require.Equal(t, tt.authorID, recipe.AuthorID().String())
				require.True(t, recipe.HasID())
			}
		})
	}
}

func TestEntity_Equals(t *testing.T) {
	ing1, _ := valueobject.NewIngredient("Flour", 500, valueobject.UnitGram)
	ing2, _ := valueobject.NewIngredient("Eggs", 3, valueobject.UnitPcs)
	validIngredients := []valueobject.Ingredient{ing1, ing2}

	step1, _ := valueobject.NewStep(1, "Mix ingredients", 5)
	step2, _ := valueobject.NewStep(2, "Bake", 30)
	validSteps := []valueobject.Step{step1, step2}

	tag1, _ := valueobject.NewTag("italian")
	tag2, _ := valueobject.NewTag("pasta")
	validTags := []valueobject.Tag{tag1, tag2}

	tests := []struct {
		name        string
		makeRecipes func() (*entity.Recipe, *entity.Recipe)
		wantEqual   bool
	}{
		{
			name: "equal entities",
			makeRecipes: func() (*entity.Recipe, *entity.Recipe) {
				recipe1, _ := entity.NewRecipe(
					"id123",
					"Pasta Carbonara",
					"Classic Italian pasta dish with eggs and cheese",
					validIngredients,
					validSteps,
					30,
					4,
					validTags,
					"author123",
				)
				recipe2, _ := entity.NewRecipe(
					"id123",
					"Pasta Carbonara",
					"Classic Italian pasta dish with eggs and cheese",
					validIngredients,
					validSteps,
					30,
					4,
					validTags,
					"author123",
				)
				return recipe1, recipe2
			},
			wantEqual: true,
		},
		{
			name: "not equal entities",
			makeRecipes: func() (*entity.Recipe, *entity.Recipe) {
				recipe1, _ := entity.NewRecipe(
					"id123",
					"Pasta Carbonara",
					"Classic Italian pasta dish with eggs and cheese",
					validIngredients,
					validSteps,
					30,
					4,
					validTags,
					"author123",
				)
				recipe2, _ := entity.NewRecipe(
					"id456",
					"Pasta Carbonara",
					"Classic Italian pasta dish with eggs and cheese",
					validIngredients,
					validSteps,
					30,
					4,
					validTags,
					"author123",
				)
				return recipe1, recipe2
			},
			wantEqual: false,
		},
		{
			name: "not equal entities - nil",
			makeRecipes: func() (*entity.Recipe, *entity.Recipe) {
				recipe1, _ := entity.NewRecipe(
					"id123",
					"Pasta Carbonara",
					"Classic Italian pasta dish with eggs and cheese",
					validIngredients,
					validSteps,
					30,
					4,
					validTags,
					"author123",
				)
				return recipe1, nil
			},
			wantEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe1, recipe2 := tt.makeRecipes()

			if tt.wantEqual {
				require.True(t, recipe1.Equals(recipe2))
			} else {
				require.False(t, recipe1.Equals(recipe2))
			}
		})
	}
}

func TestRecipe_UpdateTitle(t *testing.T) {
	recipe := fixtures.NewRecipe(t, "Test Recipe")

	t.Run("new title", func(t *testing.T) {
		newTitle, _ := valueobject.NewTitle("")

		recipe.UpdateTitle(newTitle)
		require.Equal(t, newTitle.String(), recipe.Title().String())
	})
}

func TestRecipe_UpdateDescription(t *testing.T) {
	recipe := fixtures.NewRecipe(t, "Test Recipe")

	t.Run("new description", func(t *testing.T) {
		newDescription, _ := valueobject.NewDescription("")

		recipe.UpdateDescription(newDescription)
		require.Equal(t, newDescription.String(), recipe.Description().String())
	})
}

func TestRecipe_AddIngredient(t *testing.T) {
	recipe := fixtures.NewRecipe(t, "Test Recipe")
	ingredients := recipe.Ingredients()

	t.Run("existing ingredient", func(t *testing.T) {
		ingredient, _ := valueobject.NewIngredient(ingredients[0].Name(), 500, "g")

		err := recipe.AddIngredient(ingredient)
		require.Error(t, err)
		require.ErrorIs(t, err, entity.ErrDuplicateIngredient)
		require.Len(t, recipe.Ingredients(), len(ingredients))
	})

	t.Run("new ingredient", func(t *testing.T) {
		ingredient, _ := valueobject.NewIngredient("Milk", 500, "ml")

		recipe.AddIngredient(ingredient)
		require.Len(t, recipe.Ingredients(), len(ingredients)+1)
	})
}

func TestRecipe_AddStep(t *testing.T) {
	recipe := fixtures.NewRecipe(t, "Test Recipe")
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
