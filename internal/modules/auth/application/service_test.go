package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"recipes-desk/internal/modules/auth/application"
	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/application/ports/in"
	"recipes-desk/internal/modules/auth/domain"
)

// mockUserRepository is a mock implementation of domain.Repository.
type mockUserRepository struct {
	mock.Mock
}

var _ domain.UsersRepository = (*mockUserRepository)(nil)

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// mockPasswordService is a mock implementation of passwordService.
type mockPasswordService struct {
	mock.Mock
}

var _ in.PasswordService = (*mockPasswordService)(nil)

func (m *mockPasswordService) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordService) Verify(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}

// mockTokenService is a mock implementation of tokenService.
type mockTokenService struct {
	mock.Mock
}

var _ in.TokenService = (*mockTokenService)(nil)

func (m *mockTokenService) GenerateAccessToken(userID string) (*dto.TokenResult, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TokenResult), args.Error(1)
}

func (m *mockTokenService) ValidateAccessToken(tokenString string) (*dto.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Claims), args.Error(1)
}

func (m *mockTokenService) GenerateRefreshToken(
	ctx context.Context,
	userID string,
) (*dto.TokenResult, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TokenResult), args.Error(1)
}

func (m *mockTokenService) ValidateRefreshToken(
	ctx context.Context,
	plainToken string,
) (string, error) {
	args := m.Called(ctx, plainToken)
	return args.String(0), args.Error(1)
}

func (m *mockTokenService) RotateRefreshToken(
	ctx context.Context,
	oldToken string,
) (*dto.TokenResult, string, error) {
	args := m.Called(ctx, oldToken)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*dto.TokenResult), args.String(1), args.Error(2)
}

func TestService_GetByID(t *testing.T) {
	logger := zerolog.New(nil)
	userID := "507f1f77bcf86cd799439011"

	tests := []struct {
		name       string
		id         string
		mockSetup  func(*mockUserRepository)
		wantErr    error
		wantUser   bool
		wantUserID string
	}{
		{
			name: "success",
			id:   userID,
			mockSetup: func(m *mockUserRepository) {
				m.On("FindByID", mock.Anything, userID).
					Return(&domain.User{ //nolint:exhaustruct // test struct
						ID:        userID,
						Email:     "test@example.com",
						FirstName: "John",
						LastName:  "Doe",
					}, nil)
			},
			wantErr:    nil,
			wantUser:   true,
			wantUserID: userID,
		},
		{
			name: "not found",
			id:   userID,
			mockSetup: func(m *mockUserRepository) {
				m.On("FindByID", mock.Anything, userID).
					Return(nil, domain.ErrNotFound)
			},
			wantErr:    domain.ErrNotFound,
			wantUser:   false,
			wantUserID: "",
		},
		{
			name: "repository error",
			id:   userID,
			mockSetup: func(m *mockUserRepository) {
				m.On("FindByID", mock.Anything, userID).
					Return(nil, errors.New("database error"))
			},
			wantErr:    errors.New("database error"),
			wantUser:   false,
			wantUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockUserRepository)
			tt.mockSetup(mockRepo)

			svc := application.NewService(mockRepo, &logger, nil, nil)
			user, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrNotFound) {
					require.ErrorIs(t, err, domain.ErrNotFound)
				}
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				if tt.wantUser {
					assert.Equal(t, tt.wantUserID, user.ID)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByEmail(t *testing.T) {
	logger := zerolog.New(nil)
	userID := "507f1f77bcf86cd799439011"
	email := "test@example.com"

	tests := []struct {
		name       string
		email      string
		mockSetup  func(*mockUserRepository)
		wantErr    error
		wantUser   bool
		wantUserID string
	}{
		{
			name:  "success",
			email: email,
			mockSetup: func(m *mockUserRepository) {
				m.On("FindByEmail", mock.Anything, email).
					Return(&domain.User{ //nolint:exhaustruct // test struct
						ID:        userID,
						Email:     email,
						FirstName: "John",
						LastName:  "Doe",
					}, nil)
			},
			wantErr:    nil,
			wantUser:   true,
			wantUserID: userID,
		},
		{
			name:  "not found",
			email: email,
			mockSetup: func(m *mockUserRepository) {
				m.On("FindByEmail", mock.Anything, email).
					Return(nil, domain.ErrNotFound)
			},
			wantErr:    domain.ErrNotFound,
			wantUser:   false,
			wantUserID: "",
		},
		{
			name:  "repository error",
			email: email,
			mockSetup: func(m *mockUserRepository) {
				m.On("FindByEmail", mock.Anything, email).
					Return(nil, errors.New("database error"))
			},
			wantErr:    errors.New("database error"),
			wantUser:   false,
			wantUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockUserRepository)
			tt.mockSetup(mockRepo)

			svc := application.NewService(mockRepo, &logger, nil, nil)
			user, err := svc.GetByEmail(context.Background(), tt.email)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrNotFound) {
					require.ErrorIs(t, err, domain.ErrNotFound)
				}
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				if tt.wantUser {
					assert.Equal(t, tt.wantUserID, user.ID)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Register(t *testing.T) {
	logger := zerolog.New(nil)

	tests := []struct {
		name       string
		params     dto.RegisterInput
		mockSetup  func(*mockUserRepository, *mockPasswordService, *mockTokenService)
		wantErr    error
		wantResult bool
	}{
		{
			name: "success",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, tok *mockTokenService) {
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)

				pwd.On("Hash", "password123").Return("hashed_password", nil)

				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(
						&domain.User{ //nolint:exhaustruct // test struct
							ID:        "507f1f77bcf86cd799439011",
							Email:     "test@example.com",
							FirstName: "John",
							LastName:  "Doe",
						}, nil)

				tok.On("GenerateAccessToken", mock.AnythingOfType("string")).
					Return(&dto.TokenResult{
						Token:     "access_token",
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil)

				tok.On("GenerateRefreshToken", mock.Anything, mock.AnythingOfType("string")).
					Return(&dto.TokenResult{
						Token:     "refresh_token",
						ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					}, nil)
			},
			wantErr:    nil,
			wantResult: true,
		},
		{
			name: "validation error - empty email",
			params: dto.RegisterInput{
				Email:     "",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			mockSetup: func(_ *mockUserRepository, _ *mockPasswordService, _ *mockTokenService) {
				// Repository should not be called
			},
			wantErr:    domain.ErrValidation,
			wantResult: false,
		},
		{
			name: "validation error - short password",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "short",
			},
			mockSetup: func(_ *mockUserRepository, _ *mockPasswordService, _ *mockTokenService) {
				// Repository should not be called
			},
			wantErr:    domain.ErrValidation,
			wantResult: false,
		},
		{
			name: "user already exists",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			mockSetup: func(repo *mockUserRepository, _ *mockPasswordService, _ *mockTokenService) {
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(true, nil)
			},
			wantErr:    domain.ErrAlreadyExists,
			wantResult: false,
		},
		{
			name: "exists check error",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			mockSetup: func(repo *mockUserRepository, _ *mockPasswordService, _ *mockTokenService) {
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").
					Return(false, errors.New("database error"))
			},
			wantErr:    errors.New("database error"),
			wantResult: false,
		},
		{
			name: "password hash error - too long",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, _ *mockTokenService) {
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
				pwd.On("Hash", "password123").Return("", bcrypt.ErrPasswordTooLong)
			},
			wantErr:    domain.ErrValidation,
			wantResult: false,
		},
		{
			name: "create user error",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, _ *mockTokenService) {
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)

				pwd.On("Hash", "password123").Return("hashed_password", nil)

				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(nil, errors.New("database error"))
			},
			wantErr:    errors.New("database error"),
			wantResult: false,
		},
		{
			name: "generate access token error",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, tok *mockTokenService) {
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)

				pwd.On("Hash", "password123").Return("hashed_password", nil)

				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(
						&domain.User{ //nolint:exhaustruct // test struct
							ID:        "507f1f77bcf86cd799439011",
							Email:     "test@example.com",
							FirstName: "John",
							LastName:  "Doe",
						}, nil,
					)

				tok.On("GenerateAccessToken", mock.AnythingOfType("string")).
					Return(nil, errors.New("token error"))
			},
			wantErr:    errors.New("token error"),
			wantResult: false,
		},
		{
			name: "generate refresh token error",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Password:  "password123",
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, tok *mockTokenService) {
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)

				pwd.On("Hash", "password123").Return("hashed_password", nil)

				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(
						&domain.User{ //nolint:exhaustruct // test struct
							ID:        "507f1f77bcf86cd799439011",
							Email:     "test@example.com",
							FirstName: "John",
							LastName:  "Doe",
						}, nil,
					)

				tok.On("GenerateAccessToken", mock.AnythingOfType("string")).
					Return(&dto.TokenResult{
						Token:     "access_token",
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil)

				tok.On("GenerateRefreshToken", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, errors.New("token error"))
			},
			wantErr:    errors.New("token error"),
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockUserRepository)
			mockPwd := new(mockPasswordService)
			mockTok := new(mockTokenService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockPwd, mockTok)
			}

			svc := application.NewService(mockRepo, &logger, mockPwd, mockTok)
			result, err := svc.Register(context.Background(), tt.params)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrValidation) ||
					errors.Is(tt.wantErr, domain.ErrAlreadyExists) {
					require.ErrorIs(t, err, tt.wantErr)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.wantResult {
					assert.NotNil(t, result.User)
					assert.NotEmpty(t, result.AccessToken)
					assert.NotEmpty(t, result.RefreshToken)
				}
			}

			mockRepo.AssertExpectations(t)
			mockPwd.AssertExpectations(t)
			mockTok.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	logger := zerolog.New(nil)
	userID := "507f1f77bcf86cd799439011"
	email := "test@example.com"
	password := "password123"
	hashedPassword := "hashed_password"

	tests := []struct {
		name       string
		params     dto.LoginInput
		mockSetup  func(*mockUserRepository, *mockPasswordService, *mockTokenService)
		wantErr    error
		wantResult bool
	}{
		{
			name: "success",
			params: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, tok *mockTokenService) {
				repo.On("FindByEmail", mock.Anything, email).
					Return(&domain.User{ //nolint:exhaustruct // test struct
						ID:       userID,
						Email:    email,
						Password: hashedPassword,
					}, nil)
				pwd.On("Verify", password, hashedPassword).Return(true)
				tok.On("GenerateAccessToken", userID).
					Return(&dto.TokenResult{
						Token:     "access_token",
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil)
				tok.On("GenerateRefreshToken", mock.Anything, userID).
					Return(&dto.TokenResult{
						Token:     "refresh_token",
						ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					}, nil)
			},
			wantErr:    nil,
			wantResult: true,
		},
		{
			name: "user not found",
			params: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockSetup: func(repo *mockUserRepository, _ *mockPasswordService, _ *mockTokenService) {
				repo.On("FindByEmail", mock.Anything, email).
					Return(nil, domain.ErrNotFound)
			},
			wantErr:    domain.ErrNotFound,
			wantResult: false,
		},
		{
			name: "repository error",
			params: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockSetup: func(repo *mockUserRepository, _ *mockPasswordService, _ *mockTokenService) {
				repo.On("FindByEmail", mock.Anything, email).
					Return(nil, errors.New("database error"))
			},
			wantErr:    errors.New("database error"),
			wantResult: false,
		},
		{
			name: "invalid password",
			params: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, _ *mockTokenService) {
				repo.On("FindByEmail", mock.Anything, email).
					Return(&domain.User{ //nolint:exhaustruct // test struct
						ID:       userID,
						Email:    email,
						Password: hashedPassword,
					}, nil)
				pwd.On("Verify", password, hashedPassword).Return(false)
			},
			wantErr:    domain.ErrInvalidCredentials,
			wantResult: false,
		},
		{
			name: "generate access token error",
			params: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, tok *mockTokenService) {
				repo.On("FindByEmail", mock.Anything, email).
					Return(&domain.User{ //nolint:exhaustruct // test struct
						ID:       userID,
						Email:    email,
						Password: hashedPassword,
					}, nil)
				pwd.On("Verify", password, hashedPassword).Return(true)
				tok.On("GenerateAccessToken", userID).
					Return(nil, errors.New("token error"))
			},
			wantErr:    errors.New("token error"),
			wantResult: false,
		},
		{
			name: "generate refresh token error",
			params: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, tok *mockTokenService) {
				repo.On("FindByEmail", mock.Anything, email).
					Return(&domain.User{ //nolint:exhaustruct // test struct
						ID:       userID,
						Email:    email,
						Password: hashedPassword,
					}, nil)
				pwd.On("Verify", password, hashedPassword).Return(true)
				tok.On("GenerateAccessToken", userID).
					Return(&dto.TokenResult{
						Token:     "access_token",
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil)
				tok.On("GenerateRefreshToken", mock.Anything, userID).
					Return(nil, errors.New("token error"))
			},
			wantErr:    errors.New("token error"),
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockUserRepository)
			mockPwd := new(mockPasswordService)
			mockTok := new(mockTokenService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockPwd, mockTok)
			}

			svc := application.NewService(mockRepo, &logger, mockPwd, mockTok)
			result, err := svc.Login(context.Background(), tt.params)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrNotFound) ||
					errors.Is(tt.wantErr, domain.ErrInvalidCredentials) {
					require.ErrorIs(t, err, tt.wantErr)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.wantResult {
					assert.NotNil(t, result.User)
					assert.NotEmpty(t, result.AccessToken)
					assert.NotEmpty(t, result.RefreshToken)
				}
			}

			mockRepo.AssertExpectations(t)
			mockPwd.AssertExpectations(t)
			mockTok.AssertExpectations(t)
		})
	}
}

func TestService_RefreshTokens(t *testing.T) {
	logger := zerolog.New(nil)
	userID := "507f1f77bcf86cd799439011"
	oldRefreshToken := "old_refresh_token"

	tests := []struct {
		name       string
		params     dto.RefreshTokensInput
		mockSetup  func(*mockTokenService)
		wantErr    error
		wantResult bool
	}{
		{
			name: "success",
			params: dto.RefreshTokensInput{
				RefreshToken: oldRefreshToken,
			},
			mockSetup: func(tok *mockTokenService) {
				tok.On("RotateRefreshToken", mock.Anything, oldRefreshToken).
					Return(&dto.TokenResult{
						Token:     "new_refresh_token",
						ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					}, userID, nil)
				tok.On("GenerateAccessToken", userID).
					Return(&dto.TokenResult{
						Token:     "new_access_token",
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil)
			},
			wantErr:    nil,
			wantResult: true,
		},
		{
			name: "rotate token error - invalid token",
			params: dto.RefreshTokensInput{
				RefreshToken: "invalid_token",
			},
			mockSetup: func(tok *mockTokenService) {
				tok.On("RotateRefreshToken", mock.Anything, "invalid_token").
					Return(nil, "", domain.ErrTokenNotFound)
			},
			wantErr:    domain.ErrTokenNotFound,
			wantResult: false,
		},
		{
			name: "rotate token error - expired token",
			params: dto.RefreshTokensInput{
				RefreshToken: "expired_token",
			},
			mockSetup: func(tok *mockTokenService) {
				tok.On("RotateRefreshToken", mock.Anything, "expired_token").
					Return(nil, "", domain.ErrExpiredToken)
			},
			wantErr:    domain.ErrExpiredToken,
			wantResult: false,
		},
		{
			name: "generate access token error",
			params: dto.RefreshTokensInput{
				RefreshToken: oldRefreshToken,
			},
			mockSetup: func(tok *mockTokenService) {
				tok.On("RotateRefreshToken", mock.Anything, oldRefreshToken).
					Return(&dto.TokenResult{
						Token:     "new_refresh_token",
						ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					}, userID, nil)
				tok.On("GenerateAccessToken", userID).
					Return(nil, errors.New("token error"))
			},
			wantErr:    errors.New("token error"),
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTok := new(mockTokenService)

			if tt.mockSetup != nil {
				tt.mockSetup(mockTok)
			}

			svc := application.NewService(nil, &logger, nil, mockTok)
			result, err := svc.RefreshTokens(context.Background(), tt.params)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrTokenNotFound) ||
					errors.Is(tt.wantErr, domain.ErrExpiredToken) {
					require.ErrorIs(t, err, tt.wantErr)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.wantResult {
					assert.NotEmpty(t, result.AccessToken)
					assert.NotEmpty(t, result.RefreshToken)
				}
			}

			mockTok.AssertExpectations(t)
		})
	}
}
