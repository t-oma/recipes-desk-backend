package application

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/recipes/application/dto"
	"recipes-desk/internal/modules/recipes/application/mapper"
	"recipes-desk/internal/modules/recipes/application/ports/in"
	"recipes-desk/internal/modules/recipes/domain/ports"
	"recipes-desk/internal/modules/recipes/domain/recipe"
	"recipes-desk/pkg/sliceutils"
)

// Service handles business logic for recipes.
type Service struct {
	repo  ports.RecipeRepository
	idGen ports.IDGenerator
	log   *zerolog.Logger
}

var _ in.RecipeService = (*Service)(nil)

func NewService(
	repo ports.RecipeRepository,
	idGen ports.IDGenerator,
	log *zerolog.Logger,
) *Service {
	return &Service{
		repo:  repo,
		idGen: idGen,
		log:   log,
	}
}

func (s *Service) Create(ctx context.Context, input dto.CreateRecipeInput) (*dto.Recipe, error) {
	ingredients, err := sliceutils.MapSliceWithErr(
		input.Ingredients,
		func(ing dto.IngredientInput) (recipe.Ingredient, error) {
			return recipe.NewIngredient(ing.Name, ing.Amount, ing.Unit)
		},
	)
	if err != nil {
		return nil, err
	}

	steps, err := sliceutils.MapSliceWithErr(
		input.Steps,
		func(step dto.StepInput) (recipe.Step, error) {
			return recipe.NewStep(step.Order, step.Description, step.Duration)
		},
	)
	if err != nil {
		return nil, err
	}

	tags, err := sliceutils.MapSliceWithErr(
		input.Tags,
		func(tag string) (recipe.Tag, error) {
			return recipe.NewTag(tag)
		},
	)
	if err != nil {
		return nil, err
	}

	recipe, err := recipe.NewEntity(
		s.idGen.Generate(),
		input.Title,
		input.Description,
		ingredients,
		steps,
		input.CookingTime,
		input.Portions,
		tags,
		input.UserID,
	)
	if err != nil {
		return nil, err
	}

	recipe, err = s.repo.Create(ctx, recipe)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to create recipe")
		return nil, err
	}

	return mapper.ToRecipeDTO(recipe), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*dto.Recipe, error) {
	recipe, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			s.log.Debug().Str("recipe_id", id).Msg("Recipe not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("recipe_id", id).Msg("Failed to get recipe")
		return nil, err
	}

	return mapper.ToRecipeDTO(recipe), nil
}

func (s *Service) GetAll(ctx context.Context) ([]dto.Recipe, error) {
	recipes, err := s.repo.FindAll(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get all recipes")
		return nil, err
	}

	dtos := make([]dto.Recipe, len(recipes))
	for i, recipe := range recipes {
		dtos[i] = *mapper.ToRecipeDTO(&recipe)
	}

	return dtos, nil
}

func (s *Service) Search(ctx context.Context, query string) ([]dto.Recipe, error) {
	if query == "" {
		return s.GetAll(ctx)
	}

	recipes, err := s.repo.Search(ctx, query)
	if err != nil {
		s.log.Error().Err(err).Str("query", query).Msg("Failed to search recipes")
		return nil, err
	}

	dtos := make([]dto.Recipe, len(recipes))
	for i, recipe := range recipes {
		dtos[i] = *mapper.ToRecipeDTO(&recipe)
	}
	return dtos, nil
}

func (s *Service) Update(
	ctx context.Context,
	id string,
	input dto.UpdateRecipeInput,
) (*dto.Recipe, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ingredients, err := sliceutils.MapSliceWithErr(
		input.Ingredients,
		func(ing dto.IngredientInput) (recipe.Ingredient, error) {
			return recipe.NewIngredient(ing.Name, ing.Amount, ing.Unit)
		},
	)
	if err != nil {
		return nil, err
	}

	steps, err := sliceutils.MapSliceWithErr(
		input.Steps,
		func(step dto.StepInput) (recipe.Step, error) {
			return recipe.NewStep(step.Order, step.Description, step.Duration)
		},
	)
	if err != nil {
		return nil, err
	}
	tags, err := sliceutils.MapSliceWithErr(
		input.Tags,
		func(tag string) (recipe.Tag, error) {
			return recipe.NewTag(tag)
		},
	)
	if err != nil {
		return nil, err
	}

	recipe, err := recipe.NewEntity(
		s.idGen.Generate(),
		input.Title,
		input.Description,
		ingredients,
		steps,
		input.CookingTime,
		input.Portions,
		tags,
		existing.AuthorID().Value(),
	)
	if err != nil {
		return nil, err
	}

	recipe, err = s.repo.Update(ctx, recipe)
	if err != nil {
		s.log.Error().Err(err).Str("recipe_id", id).Msg("Failed to update recipe")
		return nil, err
	}

	return mapper.ToRecipeDTO(recipe), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error().Err(err).Str("recipe_id", id).Msg("Failed to delete recipe")
		return err
	}

	return nil
}
