package service

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/recipes/domain"
)

// Service handles business logic for recipes.
type Service struct {
	repo domain.RecipesRepository
	log  *zerolog.Logger
}

func NewService(repo domain.RecipesRepository, log *zerolog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) Create(ctx context.Context, recipe *domain.Recipe) (*domain.Recipe, error) {
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
	return recipe, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Recipe, error) {
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
	return recipe, nil
}

func (s *Service) GetAll(ctx context.Context) ([]domain.Recipe, error) {
	recipes, err := s.repo.FindAll(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get all recipes")
		return nil, err
	}

	s.log.Debug().Int("count", len(recipes)).Msg("Retrieved all recipes")
	return recipes, nil
}

func (s *Service) Search(ctx context.Context, query string) ([]domain.Recipe, error) {
	if query == "" {
		return s.GetAll(ctx)
	}

	recipes, err := s.repo.Search(ctx, query)
	if err != nil {
		s.log.Error().Err(err).Str("query", query).Msg("Failed to search recipes")
		return nil, err
	}

	s.log.Debug().Str("query", query).Int("count", len(recipes)).Msg("Search completed")
	return recipes, nil
}

func (s *Service) Update(
	ctx context.Context,
	id string,
	recipe *domain.Recipe,
) (*domain.Recipe, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = recipe.Validate(); err != nil {
		s.log.Debug().Err(err).Msg("Recipe validation failed")
		return nil, err
	}

	// Preserve original ID and creation time
	recipe.ID = existing.ID
	recipe.CreatedAt = existing.CreatedAt

	recipe, err = s.repo.Update(ctx, recipe)
	if err != nil {
		s.log.Error().Err(err).Str("recipe_id", id).Msg("Failed to update recipe")
		return nil, err
	}

	s.log.Info().Str("recipe_id", id).Msg("Recipe updated")
	return recipe, nil
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
