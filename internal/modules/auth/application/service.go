package application

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/application/mapper"
	"recipes-desk/internal/modules/auth/application/ports/in"
	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/ports"
	"recipes-desk/internal/modules/auth/domain/valueobject"
)

// Service handles authentication business logic.
type Service struct {
	repo     ports.UserRepository
	log      *zerolog.Logger
	password in.PasswordService
	token    in.TokenService
	idGen    ports.IDGenerator
	uow      ports.UnitOfWork
}

var _ in.AuthService = (*Service)(nil)

// NewService creates a new auth service.
func NewService(
	repo ports.UserRepository,
	log *zerolog.Logger,
	password in.PasswordService,
	token in.TokenService,
	idGen ports.IDGenerator,
	uow ports.UnitOfWork,
) *Service {
	return &Service{
		repo:     repo,
		log:      log,
		password: password,
		token:    token,
		idGen:    idGen,
		uow:      uow,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (*dto.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, s.mapError(err, "GetByID")
	}

	return mapper.ToUserDTO(user), nil
}

func (s *Service) Register(ctx context.Context, params dto.RegisterInput) (*dto.AuthResult, error) {
	emailVO, err := valueobject.NewEmail(params.Email)
	if err != nil {
		return nil, s.mapError(err, "Register")
	}

	exists, err := s.repo.ExistsByEmail(ctx, emailVO.String())
	if err != nil {
		return nil, s.mapError(err, "Register")
	}
	if exists {
		return nil, s.mapError(domain.ErrUserAlreadyExists, "Register")
	}

	passwordVO, err := valueobject.NewPassword(params.Password)
	if err != nil {
		return nil, s.mapError(err, "Register")
	}

	hash, err := s.password.Hash(passwordVO.String())
	if err != nil {
		return nil, s.mapError(err, "Register")
	}

	idVO, err := valueobject.NewUserID(s.idGen.Generate())
	if err != nil {
		return nil, s.mapError(err, "Register")
	}
	firstNameVO, err := valueobject.NewFirstName(params.FirstName)
	if err != nil {
		return nil, s.mapError(err, "Register")
	}
	lastNameVO, err := valueobject.NewLastName(params.LastName)
	if err != nil {
		return nil, s.mapError(err, "Register")
	}

	var user *entity.User
	var accessResult *dto.TokenResult
	var refreshResult *dto.TokenResult
	err = s.uow.Execute(ctx, func(sesCtx context.Context) error {
		user, err = s.repo.Create(sesCtx, entity.NewUser(
			idVO,
			emailVO,
			firstNameVO,
			lastNameVO,
			valueobject.PasswordHash(hash),
		))
		if err != nil {
			return err
		}

		accessResult, err = s.token.GenerateAccessToken(user.ID().String())
		if err != nil {
			return err
		}

		refreshResult, err = s.token.GenerateRefreshToken(sesCtx, user.ID().String())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, s.mapError(err, "Register")
	}

	return &dto.AuthResult{
		User:             mapper.ToUserDTO(user),
		AccessToken:      accessResult.Token,
		AccessExpiresAt:  accessResult.ExpiresAt,
		RefreshToken:     refreshResult.Token,
		RefreshExpiresAt: refreshResult.ExpiresAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, params dto.LoginInput) (*dto.AuthResult, error) {
	var user *entity.User
	var accessResult *dto.TokenResult
	var refreshResult *dto.TokenResult
	err := s.uow.Execute(ctx, func(sesCtx context.Context) error {
		var err error
		user, err = s.repo.FindByEmail(sesCtx, params.Email)
		if err != nil {
			return err
		}

		if !s.password.Verify(params.Password, string(user.PasswordHash())) {
			return domain.ErrInvalidCredentials
		}

		accessResult, err = s.token.GenerateAccessToken(user.ID().String())
		if err != nil {
			return err
		}

		refreshResult, err = s.token.GenerateRefreshToken(sesCtx, user.ID().String())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, s.mapError(err, "Login")
	}

	return &dto.AuthResult{
		User:             mapper.ToUserDTO(user),
		AccessToken:      accessResult.Token,
		AccessExpiresAt:  accessResult.ExpiresAt,
		RefreshToken:     refreshResult.Token,
		RefreshExpiresAt: refreshResult.ExpiresAt,
	}, nil
}

func (s *Service) RefreshTokens(
	ctx context.Context,
	params dto.RefreshTokensInput,
) (*dto.RefreshTokensResult, error) {
	newRefreshResult, userID, err := s.token.RotateRefreshToken(ctx, params.RefreshToken)
	if err != nil {
		return nil, s.mapError(err, "RefreshTokens")
	}

	accessResult, err := s.token.GenerateAccessToken(userID)
	if err != nil {
		return nil, s.mapError(err, "RefreshTokens")
	}

	return &dto.RefreshTokensResult{
		AccessToken:      accessResult.Token,
		AccessExpiresAt:  accessResult.ExpiresAt,
		RefreshToken:     newRefreshResult.Token,
		RefreshExpiresAt: newRefreshResult.ExpiresAt,
	}, nil
}

// mapError maps domain and infrastructure errors to application-level errors.
// This ensures that only safe, client-facing errors are exposed.
func (s *Service) mapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, bcrypt.ErrPasswordTooLong):
		return valueobject.ErrPasswordTooLong

	// Domain errors that are safe to pass through
	case errors.Is(err, domain.ErrValidation):
		return err
	case errors.Is(err, domain.ErrNotFound):
		return err
	case errors.Is(err, domain.ErrForbidden):
		return err
	case errors.Is(err, domain.ErrConflict):
		return err
	case errors.Is(err, domain.ErrUnauthorized):
		return err
	case errors.Is(err, domain.ErrInvalidCredentials):
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
