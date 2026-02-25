package service

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"recipes-desk/internal/modules/auth/domain"
)

type passwordService interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

type tokenService interface {
	GenerateToken(userID, email string) (*TokenResult, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type Service struct {
	repo     domain.Repository
	log      *zerolog.Logger
	password passwordService
	token    tokenService
}

func NewService(
	repo domain.Repository,
	log *zerolog.Logger,
	password passwordService,
	token tokenService,
) *Service {
	return &Service{
		repo:     repo,
		log:      log,
		password: password,
		token:    token,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (*SafeUser, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Debug().Str("user_id", id).Msg("User not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("user_id", id).Msg("Failed to get user")
		return nil, err
	}

	s.log.Debug().Str("user_id", id).Msg("User retrieved")
	return SafeUserFromUser(user), nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*SafeUser, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Debug().Str("user_email", email).Msg("User not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("user_email", email).Msg("Failed to get user")
		return nil, err
	}

	s.log.Debug().Str("email", email).Msg("User retrieved")
	return SafeUserFromUser(user), nil
}

func (s *Service) Register(ctx context.Context, params *RegisterParams) (*AuthResult, error) {
	user := &domain.User{
		ID:                primitive.NilObjectID,
		Email:             params.Email,
		FirstName:         params.FirstName,
		LastName:          params.LastName,
		Password:          params.Password,
		CreatedAt:         time.Time{},
		PasswordUpdatedAt: time.Time{},
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

	if err = s.repo.Create(ctx, user); err != nil {
		s.log.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	s.log.Info().Str("email", params.Email).Msg("User created")

	tokenResult, err := s.token.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate token")
		return nil, err
	}

	return &AuthResult{
		User:            SafeUserFromUser(user),
		AccessToken:     tokenResult.AccessToken,
		AccessExpiresAt: tokenResult.AccessExpiresAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, params *LoginParams) (*AuthResult, error) {
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

	s.log.Debug().Str("email", params.Email).Msg("User logged in")

	tokenResult, err := s.token.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to generate token")
		return nil, err
	}

	return &AuthResult{
		User:            SafeUserFromUser(user),
		AccessToken:     tokenResult.AccessToken,
		AccessExpiresAt: tokenResult.AccessExpiresAt,
	}, nil
}
