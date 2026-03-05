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
)

// Service handles authentication business logic.
type Service struct {
	repo     domain.UsersRepository
	log      *zerolog.Logger
	password in.PasswordService
	token    in.TokenService
}

var _ in.AuthService = (*Service)(nil)

// NewService creates a new auth service.
func NewService(
	repo domain.UsersRepository,
	log *zerolog.Logger,
	password in.PasswordService,
	token in.TokenService,
) *Service {
	return &Service{
		repo:     repo,
		log:      log,
		password: password,
		token:    token,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (*dto.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Debug().Str("user_id", id).Msg("User not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("user_id", id).Msg("Failed to get user")
		return nil, err
	}

	return mapper.ToUserDTO(user), nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*dto.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Debug().Str("user_email", email).Msg("User not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("user_email", email).Msg("Failed to get user")
		return nil, err
	}

	return mapper.ToUserDTO(user), nil
}

func (s *Service) Register(ctx context.Context, params dto.RegisterInput) (*dto.AuthResult, error) {
	user := &domain.User{ //nolint:exhaustruct // fields set below
		ID:        "",
		Email:     params.Email,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		Password:  params.Password,
	}

	if err := user.Validate(); err != nil {
		s.log.Debug().Err(err).Msg("User validation failed")
		return nil, err
	}

	if exists, err := s.repo.ExistsByEmail(ctx, params.Email); err != nil {
		s.log.Error().Err(err).Msg("Failed to check if user exists")
		return nil, err
	} else if exists {
		s.log.Debug().Str("email", params.Email).Msg("User already exists")
		return nil, domain.ErrAlreadyExists
	}

	hash, err := s.password.Hash(params.Password)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to hash password")
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return nil, domain.ErrPasswordTooLong
		}
		return nil, err
	}
	user.Password = hash

	user, err = s.repo.Create(ctx, user)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	accessResult, err := s.token.GenerateAccessToken(user.ID)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate access token")
		return nil, err
	}

	refreshResult, err := s.token.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate refresh token")
		return nil, err
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
	user, err := s.repo.FindByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Debug().Str("email", params.Email).Msg("User not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("email", params.Email).Msg("Failed to get user")
		return nil, err
	}

	if !s.password.Verify(params.Password, user.Password) {
		s.log.Debug().Str("email", params.Email).Msg("Invalid password")
		return nil, domain.ErrInvalidCredentials
	}

	accessResult, err := s.token.GenerateAccessToken(user.ID)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate access token")
		return nil, err
	}

	refreshResult, err := s.token.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate refresh token")
		return nil, err
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
		s.log.Debug().Err(err).Msg("Failed to rotate refresh token")
		return nil, err
	}

	accessResult, err := s.token.GenerateAccessToken(userID)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate access token during refresh")
		return nil, err
	}

	return &dto.RefreshTokensResult{
		AccessToken:      accessResult.Token,
		AccessExpiresAt:  accessResult.ExpiresAt,
		RefreshToken:     newRefreshResult.Token,
		RefreshExpiresAt: newRefreshResult.ExpiresAt,
	}, nil
}
