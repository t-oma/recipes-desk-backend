package application

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/tags/application/dto"
	"recipes-desk/internal/modules/tags/application/mapper"
	"recipes-desk/internal/modules/tags/domain"
	"recipes-desk/internal/modules/tags/domain/entity"
	"recipes-desk/internal/modules/tags/domain/ports"
	vo "recipes-desk/internal/modules/tags/domain/valueobject"
	"recipes-desk/pkg/pagination"
)

// Service handles business logic for tags.
type Service struct {
	repo         ports.TagRepository
	idGen        ports.IDGenerator
	log          *zerolog.Logger
	defaultLimit int
	maxLimit     int
}

// NewService creates a new tag service.
func NewService(
	repo ports.TagRepository,
	idGen ports.IDGenerator,
	log *zerolog.Logger,
	maxLimit, defaultLimit int,
) *Service {
	return &Service{
		repo:         repo,
		idGen:        idGen,
		log:          log,
		maxLimit:     maxLimit,
		defaultLimit: defaultLimit,
	}
}

// CreateTag creates a new tag if it doesn't exist.
func (s *Service) CreateTag(ctx context.Context, name string) (*dto.Tag, error) {
	exists, err := s.repo.Exists(ctx, name)
	if err != nil {
		return nil, s.mapError(err, "create tag")
	}
	if exists {
		return nil, s.mapError(domain.ErrConflict, "create tag")
	}

	tagName, err := vo.NewTagName(name)
	if err != nil {
		return nil, s.mapError(err, "create tag")
	}
	tagID, err := vo.NewTagID(s.idGen.Generate())
	if err != nil {
		return nil, s.mapError(err, "create tag")
	}
	tag, err := entity.NewTag(tagID, tagName)
	if err != nil {
		return nil, s.mapError(err, "create tag")
	}

	tag, err = s.repo.Create(ctx, tag)
	if err != nil {
		return nil, s.mapError(err, "create tag")
	}

	return mapper.ToTagDTO(tag), nil
}

// GetTagByID returns a tag by its ID.
func (s *Service) GetTagByID(ctx context.Context, id string) (*dto.Tag, error) {
	tag, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, s.mapError(err, "get tag by id")
	}
	return mapper.ToTagDTO(tag), nil
}

// SearchTags searches tags by name.
func (s *Service) SearchTags(
	ctx context.Context,
	query string,
	pagn pagination.Request,
) (*pagination.Result[dto.Tag], error) {
	s.normalizePagination(&pagn)

	tags, total, err := s.repo.Search(ctx, query, pagn.Skip(), int64(pagn.Limit))
	if err != nil {
		return nil, s.mapError(err, "search tags")
	}

	dtos := make([]dto.Tag, len(tags))
	for i, tag := range tags {
		dtos[i] = *mapper.ToTagDTO(&tag)
	}

	result := pagination.NewResult(dtos, &pagn, total)
	return &result, nil
}

// EnsureTagExists ensures a tag exists, creating it if necessary.
func (s *Service) EnsureTagExists(ctx context.Context, tagName string) error {
	_, err := s.CreateTag(ctx, tagName)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil
		}
		return s.mapError(err, "ensure tag exists")
	}
	return nil
}

// EnsureTagsExist ensures multiple tags exist.
func (s *Service) EnsureTagsExist(ctx context.Context, tagNames []string) error {
	for _, name := range tagNames {
		if err := s.EnsureTagExists(ctx, name); err != nil {
			return err
		}
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
