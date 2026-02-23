package service

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"recipes-desk/internal/modules/auth/domain"
)

type PasswordService interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

type Service struct {
	repo domain.Repository
	log  *zerolog.Logger
	ps   PasswordService
}

func NewService(repo domain.Repository, log *zerolog.Logger, ps PasswordService) *Service {
	return &Service{
		repo: repo,
		log:  log,
		ps:   ps,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.User, error) {
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
	return user.ToSafe(), nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
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
	return user.ToSafe(), nil
}

func (s *Service) Register(ctx context.Context, user *domain.User) (*domain.User, error) {
	if err := user.Validate(); err != nil {
		s.log.Debug().Err(err).Msg("User validation failed")
		return nil, err
	}

	if exists, err := s.repo.ExistsByEmail(ctx, user.Email); err != nil {
		s.log.Error().Err(err).Msg("Failed to check if user exists")
		return nil, err
	} else if exists {
		s.log.Debug().Str("email", user.Email).Msg("User already exists")
		return nil, domain.ErrAlreadyExists
	}

	hash, err := s.ps.Hash(user.Password)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to hash password")
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return nil, domain.ErrPasswordTooLong
		}
		return nil, err
	}
	user.Password = hash

	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	s.log.Info().Str("email", user.Email).Msg("User created")
	return user.ToSafe(), nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Debug().Str("email", email).Msg("User not found")
			return nil, err
		}
		s.log.Error().Err(err).Str("email", email).Msg("Failed to get user")
		return nil, err
	}

	if !s.ps.Verify(password, user.Password) {
		s.log.Debug().Str("email", email).Msg("Invalid password")
		return nil, domain.ErrInvalidCredentials
	}

	s.log.Debug().Str("email", email).Msg("User logged in")
	return user.ToSafe(), nil
}
