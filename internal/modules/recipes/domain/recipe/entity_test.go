package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/domain/recipe"
)

func TestNewEntity(t *testing.T) {
	ing1, _ := recipe.NewIngredient("Flour", 500, recipe.UnitGram)
	ing2, _ := recipe.NewIngredient("Eggs", 3, recipe.UnitPcs)
	validIngredients := []recipe.Ingredient{ing1, ing2}

	step1, _ := recipe.NewStep(1, "Mix ingredients", 5)
	step2, _ := recipe.NewStep(2, "Bake", 30)
	validSteps := []recipe.Step{step1, step2}

	tag1, _ := recipe.NewTag("italian")
	tag2, _ := recipe.NewTag("pasta")
	validTags := []recipe.Tag{tag1, tag2}

	tests := []struct {
		name        string
		id          string
		title       string
		desc        string
		ingredients []recipe.Ingredient
		steps       []recipe.Step
		cookingTime int
		portions    int
		tags        []recipe.Tag
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
			ingredients: []recipe.Ingredient{},
			steps:       validSteps,
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     recipe.ErrNoIngredients,
		},
		{
			name:        "no steps",
			id:          "id123",
			title:       "Pasta Carbonara",
			desc:        "Classic Italian pasta dish with eggs and cheese",
			ingredients: validIngredients,
			steps:       []recipe.Step{},
			cookingTime: 30,
			portions:    4,
			tags:        validTags,
			authorID:    "author123",
			wantErr:     recipe.ErrNoSteps,
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
			tags:        []recipe.Tag{},
			authorID:    "author123",
			wantErr:     recipe.ErrNoTags,
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
			wantErr:     recipe.ErrRecipeIDEmpty,
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
			wantErr:     recipe.ErrTitleEmpty,
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
			wantErr:     recipe.ErrDescriptionEmpty,
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
			wantErr:     recipe.ErrCookingNegativeTime,
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
			wantErr:     recipe.ErrPortionsTooFew,
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
			wantErr:     recipe.ErrAuthorIDEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := recipe.NewEntity(
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
			}
		})
	}
}

func TestEntity_Equals(t *testing.T) {
	ing1, _ := recipe.NewIngredient("Flour", 500, recipe.UnitGram)
	ing2, _ := recipe.NewIngredient("Eggs", 3, recipe.UnitPcs)
	validIngredients := []recipe.Ingredient{ing1, ing2}

	step1, _ := recipe.NewStep(1, "Mix ingredients", 5)
	step2, _ := recipe.NewStep(2, "Bake", 30)
	validSteps := []recipe.Step{step1, step2}

	tag1, _ := recipe.NewTag("italian")
	tag2, _ := recipe.NewTag("pasta")
	validTags := []recipe.Tag{tag1, tag2}

	tests := []struct {
		name        string
		makeRecipes func() (*recipe.Entity, *recipe.Entity)
		wantEqual   bool
	}{
		{
			name: "equal entities",
			makeRecipes: func() (*recipe.Entity, *recipe.Entity) {
				recipe1, _ := recipe.NewEntity(
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
				recipe2, _ := recipe.NewEntity(
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
			makeRecipes: func() (*recipe.Entity, *recipe.Entity) {
				recipe1, _ := recipe.NewEntity(
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
				recipe2, _ := recipe.NewEntity(
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
			makeRecipes: func() (*recipe.Entity, *recipe.Entity) {
				recipe1, _ := recipe.NewEntity(
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
