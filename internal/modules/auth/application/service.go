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
}

var _ in.AuthService = (*Service)(nil)

// NewService creates a new auth service.
func NewService(
	repo ports.UserRepository,
	log *zerolog.Logger,
	password in.PasswordService,
	token in.TokenService,
	idGen ports.IDGenerator,
) *Service {
	return &Service{
		repo:     repo,
		log:      log,
		password: password,
		token:    token,
		idGen:    idGen,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (*dto.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.log.Debug().Str("user_id", id).Msg("User not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("user_id", id).Msg("Failed to get user")
		return nil, err
	}

	return mapper.ToUserDTO(user), nil
}

func (s *Service) Register(ctx context.Context, params dto.RegisterInput) (*dto.AuthResult, error) {
	emailVO, err := valueobject.NewEmail(params.Email)
	if err != nil {
		return nil, err
	}

	var exists bool
	if exists, err = s.repo.ExistsByEmail(ctx, emailVO.String()); err != nil {
		s.log.Error().Err(err).Msg("Failed to check if user exists")
		return nil, err
	} else if exists {
		s.log.Debug().Str("email", emailVO.String()).Msg("User already exists")
		return nil, domain.ErrUserAlreadyExists
	}

	passwordVO, err := valueobject.NewPassword(params.Password)
	if err != nil {
		return nil, err
	}

	hash, err := s.password.Hash(passwordVO.String())
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to hash password")
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return nil, valueobject.ErrPasswordTooLong
		}
		return nil, err
	}

	idVO, err := valueobject.NewUserID(s.idGen.Generate())
	if err != nil {
		return nil, err
	}
	firstNameVO, err := valueobject.NewFirstName(params.FirstName)
	if err != nil {
		return nil, err
	}
	lastNameVO, err := valueobject.NewLastName(params.LastName)
	if err != nil {
		return nil, err
	}

	user, err := entity.NewUser(
		idVO,
		emailVO,
		firstNameVO,
		lastNameVO,
		valueobject.PasswordHash(hash),
	)
	if err != nil {
		return nil, err
	}

	user, err = s.repo.Create(ctx, user)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	accessResult, err := s.token.GenerateAccessToken(user.ID().String())
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate access token")
		return nil, err
	}

	refreshResult, err := s.token.GenerateRefreshToken(ctx, user.ID().String())
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
		if errors.Is(err, domain.ErrUserNotFound) {
			s.log.Debug().Str("email", params.Email).Msg("User not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("email", params.Email).Msg("Failed to get user")
		return nil, err
	}

	if !s.password.Verify(params.Password, string(user.PasswordHash())) {
		s.log.Debug().Str("email", params.Email).Msg("Invalid password")
		return nil, domain.ErrInvalidCredentials
	}

	accessResult, err := s.token.GenerateAccessToken(user.ID().String())
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate access token")
		return nil, err
	}

	refreshResult, err := s.token.GenerateRefreshToken(ctx, user.ID().String())
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
