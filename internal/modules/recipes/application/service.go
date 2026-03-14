package application

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/recipes/application/dto"
	"recipes-desk/internal/modules/recipes/application/mapper"
	"recipes-desk/internal/modules/recipes/application/ports/in"
	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/ports"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
	"recipes-desk/pkg/pagination"
	"recipes-desk/pkg/sliceutils"
)

// Service handles business logic for recipes.
type Service struct {
	repo         ports.RecipeRepository
	idGen        ports.IDGenerator
	log          *zerolog.Logger
	maxLimit     int
	defaultLimit int
}

var _ in.RecipeService = (*Service)(nil)

func NewService(
	repo ports.RecipeRepository,
	idGen ports.IDGenerator,
	log *zerolog.Logger,
	maxLimit int,
	defaultLimit int,
) *Service {
	return &Service{
		repo:         repo,
		idGen:        idGen,
		log:          log,
		maxLimit:     maxLimit,
		defaultLimit: defaultLimit,
	}
}

func (s *Service) Create(ctx context.Context, input dto.CreateRecipeInput) (*dto.Recipe, error) {
	title, err := valueobject.NewTitle(input.Title)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}
	desc, err := valueobject.NewDescription(input.Description)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}
	cookingTime, err := valueobject.NewCookingTime(input.CookingTime)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}
	portions, err := valueobject.NewPortions(input.Portions)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}
	authorID, err := valueobject.NewAuthorID(input.UserID)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}

	ingredients, err := sliceutils.MapSliceWithErr(
		input.Ingredients,
		func(ing dto.IngredientInput) (valueobject.Ingredient, error) {
			return valueobject.NewIngredient(ing.Name, ing.Amount, ing.Unit)
		},
	)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}

	steps, err := sliceutils.MapSliceWithErr(
		input.Steps,
		func(step dto.StepInput) (valueobject.Step, error) {
			return valueobject.NewStep(step.Order, step.Description, step.Duration)
		},
	)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}

	tags, err := sliceutils.MapSliceWithErr(
		input.Tags,
		valueobject.NewTag,
	)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}

	id, err := valueobject.NewRecipeID(s.idGen.Generate())
	if err != nil {
		return nil, s.mapError(err, "Create")
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
		return nil, s.mapError(err, "Create")
	}

	recipe, err = s.repo.Create(ctx, recipe)
	if err != nil {
		return nil, s.mapError(err, "Create")
	}

	return mapper.ToRecipeDTO(recipe), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*dto.Recipe, error) {
	recipe, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, s.mapError(err, "GetByID")
	}

	return mapper.ToRecipeDTO(recipe), nil
}

func (s *Service) Search(
	ctx context.Context,
	query string,
	pagn pagination.Request,
) (*pagination.Result[dto.Recipe], error) {
	s.normalizePagination(&pagn)

	var (
		recipes []entity.Recipe
		total   int64
		err     error
	)
	if query == "" {
		recipes, total, err = s.repo.FindAll(ctx, pagn.Skip(), int64(pagn.Limit))
	} else {
		recipes, total, err = s.repo.Search(ctx, query, pagn.Skip(), int64(pagn.Limit))
	}
	if err != nil {
		return nil, s.mapError(err, "Search")
	}

	dtos := make([]dto.Recipe, len(recipes))
	for i, recipe := range recipes {
		dtos[i] = *mapper.ToRecipeDTO(&recipe)
	}

	result := pagination.NewResult(dtos, &pagn, total)
	return &result, nil
}

func (s *Service) Update(
	ctx context.Context,
	userID string,
	id string,
	input dto.UpdateRecipeInput,
) (*dto.Recipe, error) {
	title, err := valueobject.NewTitle(input.Title)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}
	desc, err := valueobject.NewDescription(input.Description)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}
	cookingTime, err := valueobject.NewCookingTime(input.CookingTime)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}
	portions, err := valueobject.NewPortions(input.Portions)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}

	ingredients, err := sliceutils.MapSliceWithErr(
		input.Ingredients,
		func(ing dto.IngredientInput) (valueobject.Ingredient, error) {
			return valueobject.NewIngredient(ing.Name, ing.Amount, ing.Unit)
		},
	)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}

	steps, err := sliceutils.MapSliceWithErr(
		input.Steps,
		func(step dto.StepInput) (valueobject.Step, error) {
			return valueobject.NewStep(step.Order, step.Description, step.Duration)
		},
	)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}

	tags, err := sliceutils.MapSliceWithErr(
		input.Tags,
		valueobject.NewTag,
	)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}
	if !existing.CanBeModified(userID) {
		return nil, s.mapError(domain.ErrForbidden, "Update")
	}

	recipe, err := entity.NewRecipe(
		existing.ID(),
		title,
		desc,
		ingredients,
		steps,
		cookingTime,
		portions,
		tags,
		existing.AuthorID(),
	)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}

	recipe, err = s.repo.Update(ctx, recipe)
	if err != nil {
		return nil, s.mapError(err, "Update")
	}

	return mapper.ToRecipeDTO(recipe), nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	recipeToDelete, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return s.mapError(err, "Delete")
	}

	if !recipeToDelete.CanBeModified(userID) {
		return ErrForbidden
	}

	if err = s.repo.Delete(ctx, id); err != nil {
		return s.mapError(err, "Delete")
	}

	return nil
}

// normalizePagination normalizes pagination request parameters using the service's default limit and maximum limit.
func (s *Service) normalizePagination(req *pagination.Request) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = s.defaultLimit
	}
	if req.Limit > s.maxLimit {
		req.Limit = s.maxLimit
	}
}

// mapError maps domain and infrastructure errors to application-level errors.
// This ensures that only safe, client-facing errors are exposed.
func (s *Service) mapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	switch {
	// Domain errors that are safe to pass through
	case errors.Is(err, domain.ErrValidation):
		return err
	case errors.Is(err, domain.ErrNotFound):
		return err
	case errors.Is(err, domain.ErrForbidden):
		return err
	case errors.Is(err, domain.ErrConflict):
		return err

	// Infrastructure errors - map to safe versions
	case errors.Is(err, domain.ErrTimeout):
		s.log.Error().Err(err).Str("operation", operation).Msg("Database timeout")
		return ErrServiceUnavailable
	case errors.Is(err, domain.ErrDatabase):
		s.log.Error().Err(err).Str("operation", operation).Msg("Database error")
		return ErrInternal
	default:
		s.log.Error().Err(err).Str("operation", operation).Msg("Unexpected error")
		return ErrInternal
	}
}
