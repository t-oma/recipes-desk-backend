package application_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/auth/application"
	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/ports"
	vo "recipes-desk/internal/modules/auth/domain/valueobject"
)

type mockRefreshTokensRepository struct {
	mock.Mock
}

var _ ports.RefreshTokenRepository = (*mockRefreshTokensRepository)(nil)

func (m *mockRefreshTokensRepository) Create(
	ctx context.Context,
	token *entity.RefreshToken,
	ttl time.Duration,
) (*entity.RefreshToken, error) {
	args := m.Called(ctx, token, ttl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.RefreshToken), args.Error(1)
}

func (m *mockRefreshTokensRepository) FindByHash(
	ctx context.Context,
	tokenHash string,
) (*entity.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.RefreshToken), args.Error(1)
}

func (m *mockRefreshTokensRepository) DeleteByHash(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func createTestRefreshToken(
	id, userID, tokenHash string,
	createdAt, expiresAt time.Time,
) *entity.RefreshToken {
	idVO, _ := vo.NewRefreshTokenID(id)
	userIDVO, _ := vo.NewUserID(userID)
	tokenHashVO := vo.TokenHash(tokenHash)
	token := entity.NewRefreshToken(idVO, userIDVO, tokenHashVO)
	token.RestoreFromPersistence(createdAt, expiresAt)
	return token
}

func TestTokenService_GenerateAccessToken(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-hs256"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour

	tests := []struct {
		name      string
		userID    string
		wantErr   bool
		checkFunc func(t *testing.T, result *dto.TokenResult)
	}{
		{
			name:    "success",
			userID:  "507f1f77bcf86cd799439011",
			wantErr: false,
			checkFunc: func(t *testing.T, result *dto.TokenResult) {
				t.Helper()
				assert.NotEmpty(t, result.Token)
				assert.True(t, result.ExpiresAt.After(time.Now()))
				assert.True(t, result.ExpiresAt.Before(time.Now().Add(accessTTL+time.Minute)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRefreshTokensRepository)
			mockID := new(mockIDGenerator)
			logger := zerolog.New(nil)
			tokenService := application.NewTokenService(
				mockRepo,
				mockID,
				nil,
				&logger,
				secret,
				accessTTL,
				refreshTTL,
			)

			result, err := tokenService.GenerateAccessToken(tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func generateToken(
	t *testing.T,
	secret any,
	method jwt.SigningMethod,
	userID string,
	ttl time.Duration,
) (string, error) {
	t.Helper()

	token, _, err := application.GenerateToken(secret, method, userID, ttl)
	return token, err
}

func TestTokenService_ValidateAccessToken(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-hs256"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour
	userID := "507f1f77bcf86cd799439011"

	tests := []struct {
		name       string
		generateFn func() (string, error)
		wantErr    error
		wantUserID string
	}{
		{
			name: "valid token",
			generateFn: func() (string, error) {
				return generateToken(
					t,
					[]byte(secret),
					jwt.SigningMethodHS256,
					userID,
					accessTTL,
				)
			},
			wantErr:    nil,
			wantUserID: userID,
		},
		{
			name: "expired token",
			generateFn: func() (string, error) {
				return generateToken(
					t,
					[]byte(secret),
					jwt.SigningMethodHS256,
					userID,
					-time.Hour,
				)
			},
			wantErr:    domain.ErrTokenExpired,
			wantUserID: "",
		},
		{
			name: "invalid signature",
			generateFn: func() (string, error) {
				return generateToken(
					t,
					[]byte("wrong-secret"),
					jwt.SigningMethodHS256,
					userID,
					accessTTL,
				)
			},
			wantErr:    domain.ErrTokenInvalid,
			wantUserID: "",
		},
		{
			name: "malformed token",
			generateFn: func() (string, error) {
				return "invalid.token.string", nil
			},
			wantErr:    domain.ErrTokenInvalid,
			wantUserID: "",
		},
		{
			name: "empty token",
			generateFn: func() (string, error) {
				return "", nil
			},
			wantErr:    domain.ErrTokenInvalid,
			wantUserID: "",
		},
		{
			name: "token with invalid method",
			generateFn: func() (string, error) {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					t.Fatal(err)
				}

				return generateToken(
					t,
					key,
					jwt.SigningMethodRS256,
					userID,
					accessTTL,
				)
			},
			wantErr:    domain.ErrTokenInvalid,
			wantUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRefreshTokensRepository)
			mockID := new(mockIDGenerator)
			logger := zerolog.New(nil)
			tokenService := application.NewTokenService(
				mockRepo,
				mockID,
				nil,
				&logger,
				secret,
				accessTTL,
				refreshTTL,
			)

			token, err := tt.generateFn()
			require.NoError(t, err)

			claims, validateErr := tokenService.ValidateAccessToken(token)

			if tt.wantErr != nil {
				require.Error(t, validateErr)
				require.ErrorIs(t, validateErr, tt.wantErr)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, validateErr)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.wantUserID, claims.UserID)
			}
		})
	}
}

func TestTokenService_GenerateRefreshToken(t *testing.T) {
	secret := "test-secret"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour
	userID := "507f1f77bcf86cd799439011"

	tests := []struct {
		name      string
		mockSetup func(*mockRefreshTokensRepository, *mockIDGenerator)
		wantErr   bool
		checkFunc func(t *testing.T, result *dto.TokenResult)
	}{
		{
			name: "success",
			mockSetup: func(m *mockRefreshTokensRepository, idGen *mockIDGenerator) {
				idGen.On("Generate").Return("507f1f77bcf86cd799439011")
				m.On("Create", mock.Anything, mock.AnythingOfType("*entity.RefreshToken"), mock.AnythingOfType("time.Duration")).
					Return(
						createTestRefreshToken(
							"507f1f77bcf86cd799439011",
							userID,
							"somehash",
							time.Now().Add(-time.Hour),
							time.Now().Add(refreshTTL),
						),
						nil,
					)
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *dto.TokenResult) {
				t.Helper()
				assert.NotEmpty(t, result.Token)
				assert.True(t, result.ExpiresAt.After(time.Now()))
				assert.True(t, result.ExpiresAt.Before(time.Now().Add(refreshTTL+time.Hour)))
			},
		},
		{
			name: "repository error",
			mockSetup: func(m *mockRefreshTokensRepository, idGen *mockIDGenerator) {
				idGen.On("Generate").Return("507f1f77bcf86cd799439011").Once()
				m.On("Create", mock.Anything, mock.AnythingOfType("*entity.RefreshToken"), mock.AnythingOfType("time.Duration")).
					Return(nil, errors.New("database error"))
			},
			wantErr:   true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRefreshTokensRepository)
			mockID := new(mockIDGenerator)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockID)
			}

			mockUOW := new(mockUnitOfWork)
			mockUOW.On("Execute", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)
			logger := zerolog.New(nil)
			tokenService := application.NewTokenService(
				mockRepo,
				mockID,
				mockUOW,
				&logger,
				secret,
				accessTTL,
				refreshTTL,
			)
			result, err := tokenService.GenerateRefreshToken(context.Background(), userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
			mockID.AssertExpectations(t)
		})
	}
}

func TestTokenService_ValidateRefreshToken(t *testing.T) {
	secret := "test-secret"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour
	userID := "507f1f77bcf86cd799439011"
	tokenHash := "test-token-hash"

	tests := []struct {
		name       string
		plainToken string
		mockSetup  func(*mockRefreshTokensRepository)
		wantErr    error
		wantUserID string
	}{
		{
			name:       "valid token",
			plainToken: "valid-plain-token",
			mockSetup: func(m *mockRefreshTokensRepository) {
				m.On("FindByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(
						createTestRefreshToken(
							"test-id",
							userID,
							tokenHash,
							time.Now().Add(-time.Hour),
							time.Now().Add(time.Hour),
						),
						nil,
					)
			},
			wantErr:    nil,
			wantUserID: userID,
		},
		{
			name:       "token not found",
			plainToken: "non-existent-token",
			mockSetup: func(m *mockRefreshTokensRepository) {
				m.On("FindByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, domain.ErrTokenNotFound)
			},
			wantErr:    domain.ErrUnauthorized,
			wantUserID: "",
		},
		{
			name:       "token expired",
			plainToken: "expired-token",
			mockSetup: func(m *mockRefreshTokensRepository) {
				m.On("FindByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(
						createTestRefreshToken(
							"test-id",
							userID,
							tokenHash,
							time.Now().Add(-2*time.Hour),
							time.Now().Add(-time.Hour),
						),
						nil,
					)
			},
			wantErr:    domain.ErrTokenExpired,
			wantUserID: "",
		},
		{
			name:       "repository error",
			plainToken: "error-token",
			mockSetup: func(m *mockRefreshTokensRepository) {
				m.On("FindByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, domain.ErrDatabase)
			},
			wantErr:    application.ErrInternal,
			wantUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRefreshTokensRepository)
			mockID := new(mockIDGenerator)
			tt.mockSetup(mockRepo)

			logger := zerolog.New(nil)
			tokenService := application.NewTokenService(
				mockRepo,
				mockID,
				nil,
				&logger,
				secret,
				accessTTL,
				refreshTTL,
			)
			resultUserID, err := tokenService.ValidateRefreshToken(
				context.Background(),
				tt.plainToken,
			)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, tt.wantUserID, resultUserID)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUserID, resultUserID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenService_RotateRefreshToken(t *testing.T) {
	secret := "test-secret"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour
	userID := "507f1f77bcf86cd799439011"

	tests := []struct {
		name          string
		oldToken      string
		mockSetup     func(*mockRefreshTokensRepository, *mockIDGenerator, *mockUnitOfWork)
		wantErr       error
		checkNewToken bool
	}{
		{
			name:     "success",
			oldToken: "valid-old-token",
			mockSetup: func(m *mockRefreshTokensRepository, mID *mockIDGenerator, mUOW *mockUnitOfWork) {
				mUOW.On("Execute", mock.Anything, mock.Anything).Return(nil)
				m.On("FindByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(
						createTestRefreshToken(
							"test-id",
							userID,
							"old-hash",
							time.Now().Add(-time.Hour),
							time.Now().Add(time.Hour),
						),
						nil,
					).Once()

				m.On("DeleteByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(nil).Once()

				mID.On("Generate").Return("507f1f77bcf86cd799439011")

				m.On("Create", mock.Anything, mock.AnythingOfType("*entity.RefreshToken"), mock.AnythingOfType("time.Duration")).
					Return(
						createTestRefreshToken(
							"507f1f77bcf86cd799439011",
							userID,
							"new-hash",
							time.Now().Add(-time.Hour),
							time.Now().Add(refreshTTL),
						),
						nil,
					).
					Once()
			},
			wantErr:       nil,
			checkNewToken: true,
		},
		{
			name:     "invalid old token",
			oldToken: "invalid-token",
			mockSetup: func(m *mockRefreshTokensRepository, _ *mockIDGenerator, mUOW *mockUnitOfWork) {
				mUOW.On("Execute", mock.Anything, mock.Anything).Return(nil)
				m.On("FindByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, domain.ErrTokenNotFound).Once()
			},
			wantErr:       domain.ErrUnauthorized,
			checkNewToken: false,
		},
		{
			name:     "delete old token error",
			oldToken: "valid-old-token",
			mockSetup: func(m *mockRefreshTokensRepository, _ *mockIDGenerator, mUOW *mockUnitOfWork) {
				mUOW.On("Execute", mock.Anything, mock.Anything).Return(nil)
				m.On("FindByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(
						createTestRefreshToken(
							"test-id",
							userID,
							"old-hash",
							time.Now().Add(-time.Hour),
							time.Now().Add(time.Hour),
						),
						nil,
					).Once()
				m.On("DeleteByHash", mock.Anything, mock.AnythingOfType("string")).
					Return(domain.ErrDatabase).Once()
			},
			wantErr:       application.ErrInternal,
			checkNewToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRefreshTokensRepository)
			mockID := new(mockIDGenerator)
			mockUOW := new(mockUnitOfWork)
			tt.mockSetup(mockRepo, mockID, mockUOW)

			logger := zerolog.New(nil)
			tokenService := application.NewTokenService(
				mockRepo,
				mockID,
				mockUOW,
				&logger,
				secret,
				accessTTL,
				refreshTTL,
			)
			newToken, resultUserID, err := tokenService.RotateRefreshToken(
				context.Background(),
				tt.oldToken,
			)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, newToken)
				assert.Empty(t, resultUserID)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, newToken)
				assert.Equal(t, userID, resultUserID)
				if tt.checkNewToken {
					assert.NotEmpty(t, newToken.Token)
					assert.True(t, newToken.ExpiresAt.After(time.Now()))
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
