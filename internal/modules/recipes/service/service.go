package service

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/recipes/domain"
)

// RecipeService defines the service interface.
type RecipeService interface {
	Create(ctx context.Context, recipe CreateRecipeInput) (*RecipeDTO, error)
	GetByID(ctx context.Context, id string) (*RecipeDTO, error)
	GetAll(ctx context.Context) ([]RecipeDTO, error)
	Search(ctx context.Context, query string) ([]RecipeDTO, error)
	Update(
		ctx context.Context,
		id string,
		recipe UpdateRecipeInput,
	) (*RecipeDTO, error)
	Delete(ctx context.Context, id string) error
}

// Service handles business logic for recipes.
type Service struct {
	repo domain.RecipesRepository
	log  *zerolog.Logger
}

var _ RecipeService = (*Service)(nil)

func NewService(repo domain.RecipesRepository, log *zerolog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func toDomainIngredients(ingredients []IngredientInput) []domain.Ingredient {
	mapped := make([]domain.Ingredient, len(ingredients))
	for i, ing := range ingredients {
		mapped[i] = domain.Ingredient{
			Name:   ing.Name,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	return mapped
}

func toDomainSteps(steps []StepInput) []domain.Step {
	mapped := make([]domain.Step, len(steps))
	for i, step := range steps {
		mapped[i] = domain.Step{
			Order:       step.Order,
			Description: step.Description,
			Duration:    step.Duration,
		}
	}
	return mapped
}

func (s *Service) Create(ctx context.Context, input CreateRecipeInput) (*RecipeDTO, error) {
	recipe := &domain.Recipe{
		ID:          "",
		Title:       input.Title,
		Description: input.Description,
		Ingredients: toDomainIngredients(input.Ingredients),
		Steps:       toDomainSteps(input.Steps),
		CookingTime: input.CookingTime,
		Portions:    input.Portions,
		Tags:        input.Tags,
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	}

	if err := recipe.Validate(); err != nil {
		s.log.Debug().Err(err).Msg("Recipe validation failed")
		return nil, err
	}

	recipe, err := s.repo.Create(ctx, recipe)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to create recipe")
		return nil, err
	}

	s.log.Info().Str("recipe_id", recipe.ID).Msg("Recipe created")
	return toRecipeDTO(recipe), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*RecipeDTO, error) {
	recipe, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Debug().Str("recipe_id", id).Msg("Recipe not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("recipe_id", id).Msg("Failed to get recipe")
		return nil, err
	}

	s.log.Debug().Str("recipe_id", id).Msg("Recipe retrieved")
	return toRecipeDTO(recipe), nil
}

func (s *Service) GetAll(ctx context.Context) ([]RecipeDTO, error) {
	recipes, err := s.repo.FindAll(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get all recipes")
		return nil, err
	}

	s.log.Debug().Int("count", len(recipes)).Msg("Retrieved all recipes")
	dtos := make([]RecipeDTO, len(recipes))
	for i, recipe := range recipes {
		dtos[i] = *toRecipeDTO(&recipe)
	}

	return dtos, nil
}

func (s *Service) Search(ctx context.Context, query string) ([]RecipeDTO, error) {
	if query == "" {
		return s.GetAll(ctx)
	}

	recipes, err := s.repo.Search(ctx, query)
	if err != nil {
		s.log.Error().Err(err).Str("query", query).Msg("Failed to search recipes")
		return nil, err
	}

	s.log.Debug().Str("query", query).Int("count", len(recipes)).Msg("Search completed")
	dtos := make([]RecipeDTO, len(recipes))
	for i, recipe := range recipes {
		dtos[i] = *toRecipeDTO(&recipe)
	}
	return dtos, nil
}

func (s *Service) Update(
	ctx context.Context,
	id string,
	input UpdateRecipeInput,
) (*RecipeDTO, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	recipe := &domain.Recipe{
		ID:          existing.ID, // Preserve original ID and creation time
		Title:       input.Title,
		Description: input.Description,
		Ingredients: toDomainIngredients(input.Ingredients),
		Steps:       toDomainSteps(input.Steps),
		CookingTime: input.CookingTime,
		Portions:    input.Portions,
		Tags:        input.Tags,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   time.Time{},
	}

	if err = recipe.Validate(); err != nil {
		s.log.Debug().Err(err).Msg("Recipe validation failed")
		return nil, err
	}

	recipe, err = s.repo.Update(ctx, recipe)
	if err != nil {
		s.log.Error().Err(err).Str("recipe_id", id).Msg("Failed to update recipe")
		return nil, err
	}

	s.log.Info().Str("recipe_id", id).Msg("Recipe updated")
	return toRecipeDTO(recipe), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error().Err(err).Str("recipe_id", id).Msg("Failed to delete recipe")
		return err
	}

	s.log.Info().Str("recipe_id", id).Msg("Recipe deleted")
	return nil
}
