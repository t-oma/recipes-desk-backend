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
	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/fixtures"
	"recipes-desk/internal/modules/auth/domain/ports"
	"recipes-desk/internal/modules/auth/domain/valueobject"
)

// mockUserRepository is a mock implementation of ports.UserRepository.
type mockUserRepository struct {
	mock.Mock
}

var _ ports.UserRepository = (*mockUserRepository)(nil)

func (m *mockUserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
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

// mockIDGenerator is a mock implementation of ports.IDGenerator.
type mockIDGenerator struct {
	mock.Mock
}

var _ ports.IDGenerator = (*mockIDGenerator)(nil)

func (m *mockIDGenerator) Generate() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockIDGenerator) Validate(id string) error {
	args := m.Called(id)
	return args.Error(0)
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
				user := fixtures.NewUser(t, "test@example.com")
				id, _ := valueobject.NewUserID(userID)
				user.AssignID(id)
				m.On("FindByID", mock.Anything, userID).
					Return(user, nil)
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
					Return(nil, domain.ErrUserNotFound)
			},
			wantErr:    domain.ErrUserNotFound,
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

			svc := application.NewService(mockRepo, &logger, nil, nil, nil)
			user, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrUserNotFound) {
					require.ErrorIs(t, err, domain.ErrUserNotFound)
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
		mockSetup  func(*mockUserRepository, *mockPasswordService, *mockTokenService, *mockIDGenerator)
		wantErr    error
		wantResult bool
	}{
		{
			name: "success",
			params: dto.RegisterInput{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, tok *mockTokenService, idGen *mockIDGenerator) {
				userID := "user-id-123"
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").
					Return(false, nil).Once()
				pwd.On("Hash", "password123").
					Return("hashedpassword", nil).Once()
				idGen.On("Generate").Return(userID).Once()
				user := fixtures.NewUserWithOptions(t,
					fixtures.WithEmail("test@example.com"),
					fixtures.WithFirstName("John"),
					fixtures.WithLastName("Doe"),
				)
				userIDVO, _ := valueobject.NewUserID(userID)
				user.AssignID(userIDVO)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).
					Return(user, nil).Once()
				tok.On("GenerateAccessToken", userID).
					Return(&dto.TokenResult{Token: "access-token", ExpiresAt: time.Now().Add(time.Hour)}, nil).
					Once()
				tok.On("GenerateRefreshToken", mock.Anything, userID).
					Return(&dto.TokenResult{Token: "refresh-token", ExpiresAt: time.Now().Add(24 * time.Hour)}, nil).
					Once()
			},
			wantErr:    nil,
			wantResult: true,
		},
		{
			name: "email already exists",
			params: dto.RegisterInput{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(repo *mockUserRepository, _ *mockPasswordService, _ *mockTokenService, _ *mockIDGenerator) {
				repo.On("ExistsByEmail", mock.Anything, "existing@example.com").
					Return(true, nil).Once()
			},
			wantErr:    domain.ErrUserAlreadyExists,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockUserRepository)
			mockPwd := new(mockPasswordService)
			mockTok := new(mockTokenService)
			mockID := new(mockIDGenerator)
			tt.mockSetup(mockRepo, mockPwd, mockTok, mockID)

			svc := application.NewService(mockRepo, &logger, mockPwd, mockTok, mockID)
			result, err := svc.Register(context.Background(), tt.params)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.wantResult {
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
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

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
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, tok *mockTokenService) {
				user := fixtures.NewUserWithOptions(t,
					fixtures.WithEmail("test@example.com"),
					fixtures.WithPassword(string(hashedPassword)),
				)
				id, _ := valueobject.NewUserID("user-id-123")
				user.AssignID(id)
				repo.On("FindByEmail", mock.Anything, "test@example.com").
					Return(user, nil).Once()
				pwd.On("Verify", "password123", string(hashedPassword)).
					Return(true).Once()
				tok.On("GenerateAccessToken", "user-id-123").
					Return(&dto.TokenResult{Token: "access-token", ExpiresAt: time.Now().Add(time.Hour)}, nil).
					Once()
				tok.On("GenerateRefreshToken", mock.Anything, "user-id-123").
					Return(&dto.TokenResult{Token: "refresh-token", ExpiresAt: time.Now().Add(24 * time.Hour)}, nil).
					Once()
			},
			wantErr:    nil,
			wantResult: true,
		},
		{
			name: "user not found",
			params: dto.LoginInput{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockSetup: func(repo *mockUserRepository, _ *mockPasswordService, _ *mockTokenService) {
				repo.On("FindByEmail", mock.Anything, "nonexistent@example.com").
					Return(nil, domain.ErrUserNotFound).Once()
			},
			wantErr:    domain.ErrUserNotFound,
			wantResult: false,
		},
		{
			name: "invalid password",
			params: dto.LoginInput{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(repo *mockUserRepository, pwd *mockPasswordService, _ *mockTokenService) {
				user := fixtures.NewUserWithOptions(t,
					fixtures.WithEmail("test@example.com"),
					fixtures.WithPassword(string(hashedPassword)),
				)
				id, _ := valueobject.NewUserID("user-id-123")
				user.AssignID(id)
				repo.On("FindByEmail", mock.Anything, "test@example.com").
					Return(user, nil).Once()
				pwd.On("Verify", "wrongpassword", string(hashedPassword)).
					Return(false).Once()
			},
			wantErr:    domain.ErrInvalidCredentials,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockUserRepository)
			mockPwd := new(mockPasswordService)
			mockTok := new(mockTokenService)
			tt.mockSetup(mockRepo, mockPwd, mockTok)

			svc := application.NewService(mockRepo, &logger, mockPwd, mockTok, nil)
			result, err := svc.Login(context.Background(), tt.params)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.wantResult {
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
				RefreshToken: "valid-refresh-token",
			},
			mockSetup: func(tok *mockTokenService) {
				tok.On("RotateRefreshToken", mock.Anything, "valid-refresh-token").
					Return(&dto.TokenResult{Token: "new-refresh-token", ExpiresAt: time.Now().Add(24 * time.Hour)}, "user-id-123", nil).
					Once()
				tok.On("GenerateAccessToken", "user-id-123").
					Return(&dto.TokenResult{Token: "new-access-token", ExpiresAt: time.Now().Add(time.Hour)}, nil).
					Once()
			},
			wantErr:    nil,
			wantResult: true,
		},
		{
			name: "invalid refresh token",
			params: dto.RefreshTokensInput{
				RefreshToken: "invalid-token",
			},
			mockSetup: func(tok *mockTokenService) {
				tok.On("RotateRefreshToken", mock.Anything, "invalid-token").
					Return(nil, "", domain.ErrTokenInvalid).Once()
			},
			wantErr:    domain.ErrTokenInvalid,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTok := new(mockTokenService)
			tt.mockSetup(mockTok)

			svc := application.NewService(nil, &logger, nil, mockTok, nil)
			result, err := svc.RefreshTokens(context.Background(), tt.params)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
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
