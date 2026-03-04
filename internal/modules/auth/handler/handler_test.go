package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/handler"
	"recipes-desk/internal/modules/auth/service"
)

// mockAuthService is a mock implementation of handler.AuthService.
type mockAuthService struct {
	mock.Mock
}

var _ handler.AuthService = (*mockAuthService)(nil)

func (m *mockAuthService) GetByID(ctx context.Context, id string) (*service.UserDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UserDTO), args.Error(1)
}

func (m *mockAuthService) GetByEmail(ctx context.Context, email string) (*service.UserDTO, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UserDTO), args.Error(1)
}

func (m *mockAuthService) Register(
	ctx context.Context,
	params *service.RegisterInput,
) (*service.AuthResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResult), args.Error(1)
}

func (m *mockAuthService) Login(
	ctx context.Context,
	params *service.LoginInput,
) (*service.AuthResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResult), args.Error(1)
}

func (m *mockAuthService) RefreshTokens(
	ctx context.Context,
	input *service.RefreshTokensInput,
) (*service.RefreshTokensResult, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RefreshTokensResult), args.Error(1)
}

func setupTest() (*gin.Engine, *mockAuthService, *handler.Handler) {
	gin.SetMode(gin.TestMode)
	logger := zerolog.New(nil)

	mockSvc := new(mockAuthService)
	h := handler.NewHandler(mockSvc, &logger)

	router := gin.New()

	return router, mockSvc, h
}

func TestHandler_Register(t *testing.T) {
	userID := "507f1f77bcf86cd799439011"

	tests := []struct {
		name           string
		body           map[string]any
		mockSetup      func(*mockAuthService)
		wantStatusCode int
		wantCookies    int
	}{
		{
			name: "success",
			body: map[string]any{
				"email":     "test@example.com",
				"firstName": "John",
				"lastName":  "Doe",
				"password":  "password123",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*service.RegisterParams")).
					Return(&service.AuthResult{
						User: &service.UserDTO{ //nolint:exhaustruct // test struct
							ID:        userID,
							Email:     "test@example.com",
							FirstName: "John",
							LastName:  "Doe",
						},
						AccessToken:      "access_token",
						AccessExpiresAt:  time.Now().Add(time.Hour),
						RefreshToken:     "refresh_token",
						RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					}, nil)
			},
			wantStatusCode: http.StatusCreated,
			wantCookies:    2,
		},
		{
			name: "invalid json",
			body: nil,
			mockSetup: func(_ *mockAuthService) {
				// Service should not be called
			},
			wantStatusCode: http.StatusBadRequest,
			wantCookies:    0,
		},
		{
			name: "validation error",
			body: map[string]any{
				"email":     "invalid-email",
				"firstName": "John",
				"lastName":  "Doe",
				"password":  "password123",
			},
			mockSetup: func(_ *mockAuthService) {
				// Service should not be called
			},
			wantStatusCode: http.StatusBadRequest,
			wantCookies:    0,
		},
		{
			name: "user already exists",
			body: map[string]any{
				"email":     "test@example.com",
				"firstName": "John",
				"lastName":  "Doe",
				"password":  "password123",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*service.RegisterParams")).
					Return(nil, domain.ErrAlreadyExists)
			},
			wantStatusCode: http.StatusConflict,
			wantCookies:    0,
		},
		{
			name: "service error",
			body: map[string]any{
				"email":     "test@example.com",
				"firstName": "John",
				"lastName":  "Doe",
				"password":  "password123",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*service.RegisterParams")).
					Return(nil, errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantCookies:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			router.POST("/auth/register", h.Register)

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte(`{invalid json`)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantStatusCode == http.StatusCreated {
				var response handler.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, userID, response.User.ID)

				// Check cookies are set
				cookies := w.Result().Cookies()
				assert.Len(t, cookies, tt.wantCookies)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	userID := "507f1f77bcf86cd799439011"

	tests := []struct {
		name           string
		body           map[string]any
		mockSetup      func(*mockAuthService)
		wantStatusCode int
		wantCookies    int
	}{
		{
			name: "success",
			body: map[string]any{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Login", mock.Anything, mock.AnythingOfType("*service.LoginParams")).
					Return(&service.AuthResult{
						User: &service.UserDTO{ //nolint:exhaustruct // test struct
							ID:        userID,
							Email:     "test@example.com",
							FirstName: "John",
							LastName:  "Doe",
						},
						AccessToken:      "access_token",
						AccessExpiresAt:  time.Now().Add(time.Hour),
						RefreshToken:     "refresh_token",
						RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantCookies:    2,
		},
		{
			name: "invalid json",
			body: nil,
			mockSetup: func(_ *mockAuthService) {
				// Service should not be called
			},
			wantStatusCode: http.StatusBadRequest,
			wantCookies:    0,
		},
		{
			name: "user not found",
			body: map[string]any{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Login", mock.Anything, mock.AnythingOfType("*service.LoginParams")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantCookies:    0,
		},
		{
			name: "invalid credentials",
			body: map[string]any{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Login", mock.Anything, mock.AnythingOfType("*service.LoginParams")).
					Return(nil, domain.ErrInvalidCredentials)
			},
			wantStatusCode: http.StatusBadRequest,
			wantCookies:    0,
		},
		{
			name: "service error",
			body: map[string]any{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockSetup: func(m *mockAuthService) {
				m.On("Login", mock.Anything, mock.AnythingOfType("*service.LoginParams")).
					Return(nil, errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantCookies:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			router.POST("/auth/login", h.Login)

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte(`{invalid json`)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantStatusCode == http.StatusOK {
				var response handler.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, userID, response.User.ID)

				// Check cookies are set
				cookies := w.Result().Cookies()
				assert.Len(t, cookies, tt.wantCookies)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	tests := []struct {
		name           string
		cookie         string
		mockSetup      func(*mockAuthService)
		wantStatusCode int
		wantCookies    int
	}{
		{
			name:   "success",
			cookie: "valid_refresh_token",
			mockSetup: func(m *mockAuthService) {
				m.On("RefreshTokens", mock.Anything, mock.AnythingOfType("*service.RefreshTokensParams")).
					Return(&service.RefreshTokensResult{
						AccessToken:      "new_access_token",
						AccessExpiresAt:  time.Now().Add(time.Hour),
						RefreshToken:     "new_refresh_token",
						RefreshExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantCookies:    2,
		},
		{
			name:           "missing cookie",
			cookie:         "",
			mockSetup:      func(_ *mockAuthService) {},
			wantStatusCode: http.StatusUnauthorized,
			wantCookies:    0,
		},
		{
			name:   "invalid token",
			cookie: "invalid_token",
			mockSetup: func(m *mockAuthService) {
				m.On("RefreshTokens", mock.Anything, mock.AnythingOfType("*service.RefreshTokensParams")).
					Return(nil, domain.ErrTokenNotFound)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantCookies:    0,
		},
		{
			name:   "expired token",
			cookie: "expired_token",
			mockSetup: func(m *mockAuthService) {
				m.On("RefreshTokens", mock.Anything, mock.AnythingOfType("*service.RefreshTokensParams")).
					Return(nil, domain.ErrExpiredToken)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantCookies:    0,
		},
		{
			name:   "service error",
			cookie: "valid_token",
			mockSetup: func(m *mockAuthService) {
				m.On("RefreshTokens", mock.Anything, mock.AnythingOfType("*service.RefreshTokensParams")).
					Return(nil, errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantCookies:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			router.POST("/auth/refresh", h.Refresh)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", nil)
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.cookie,
				})
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantStatusCode == http.StatusOK {
				// Check cookies are set
				cookies := w.Result().Cookies()
				assert.Len(t, cookies, tt.wantCookies)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	router, _, h := setupTest()

	router.POST("/auth/logout", h.Logout)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
	// Set some cookies to verify they are cleared
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: "some_token",
	})
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "some_refresh_token",
	})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "logged out successfully", response["message"])

	// Check cookies are cleared (MaxAge = -1)
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 2)
	for _, cookie := range cookies {
		assert.Equal(t, -1, cookie.MaxAge)
	}
}

func TestHandler_Me(t *testing.T) {
	userID := "507f1f77bcf86cd799439011"

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*mockAuthService)
		wantStatusCode int
		wantUser       bool
	}{
		{
			name:   "success",
			userID: userID,
			mockSetup: func(m *mockAuthService) {
				m.On("GetByID", mock.Anything, userID).
					Return(&service.UserDTO{ //nolint:exhaustruct // test struct
						ID:        userID,
						Email:     "test@example.com",
						FirstName: "John",
						LastName:  "Doe",
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantUser:       true,
		},
		{
			name:           "unauthorized - no userID",
			userID:         "",
			mockSetup:      func(_ *mockAuthService) {},
			wantStatusCode: http.StatusUnauthorized,
			wantUser:       false,
		},
		{
			name:   "user not found",
			userID: userID,
			mockSetup: func(m *mockAuthService) {
				m.On("GetByID", mock.Anything, userID).
					Return(nil, domain.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantUser:       false,
		},
		{
			name:   "service error",
			userID: userID,
			mockSetup: func(m *mockAuthService) {
				m.On("GetByID", mock.Anything, userID).
					Return(nil, errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantUser:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			router.GET("/auth/me", func(c *gin.Context) {
				if tt.userID != "" {
					c.Set("userID", tt.userID)
				}
				h.Me(c)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantUser {
				var response handler.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, userID, response.ID)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
