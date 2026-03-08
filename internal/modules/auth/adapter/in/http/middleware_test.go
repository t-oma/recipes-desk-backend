package httphandler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	httphandler "recipes-desk/internal/modules/auth/adapter/in/http"
	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/domain"
)

// mockTokenValidator is a mock implementation of tokenService for middleware testing.
type mockTokenValidator struct {
	mock.Mock
}

func (m *mockTokenValidator) ValidateAccessToken(tokenString string) (*dto.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Claims), args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		mockSetup      func(*mockTokenValidator)
		wantStatusCode int
		wantUserID     string
		wantNextCalled bool
	}{
		{
			name: "success - valid token from cookie",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "valid_token",
				})
			},
			mockSetup: func(m *mockTokenValidator) {
				m.On("ValidateAccessToken", "valid_token").Return(
					&dto.Claims{UserID: "user123"}, //nolint:exhaustruct // test struct
					nil,
				)
			},
			wantStatusCode: http.StatusOK,
			wantUserID:     "user123",
			wantNextCalled: true,
		},
		{
			name: "success - valid token from Authorization header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer valid_header_token")
			},
			mockSetup: func(m *mockTokenValidator) {
				m.On("ValidateAccessToken", "valid_header_token").Return(
					&dto.Claims{UserID: "user456"}, //nolint:exhaustruct // test struct
					nil,
				)
			},
			wantStatusCode: http.StatusOK,
			wantUserID:     "user456",
			wantNextCalled: true,
		},
		{
			name:           "missing token - no cookie no header",
			setupRequest:   func(_ *http.Request) {},
			mockSetup:      func(_ *mockTokenValidator) {},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     "",
			wantNextCalled: false,
		},
		{
			name: "invalid Authorization header format - no Bearer",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Basic invalid_token")
			},
			mockSetup:      func(_ *mockTokenValidator) {},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     "",
			wantNextCalled: false,
		},
		{
			name: "invalid Authorization header format - missing token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer ")
			},
			mockSetup: func(m *mockTokenValidator) {
				m.On("ValidateAccessToken", "").
					Return(nil, domain.ErrTokenInvalid)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     "",
			wantNextCalled: false,
		},
		{
			name: "expired token",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "expired_token",
				})
			},
			mockSetup: func(m *mockTokenValidator) {
				m.On("ValidateAccessToken", "expired_token").
					Return(nil, domain.ErrTokenExpired)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     "",
			wantNextCalled: false,
		},
		{
			name: "invalid token",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "invalid_token",
				})
			},
			mockSetup: func(m *mockTokenValidator) {
				m.On("ValidateAccessToken", "invalid_token").
					Return(nil, domain.ErrTokenInvalid)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     "",
			wantNextCalled: false,
		},
		{
			name: "token validation error",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "error_token",
				})
			},
			mockSetup: func(m *mockTokenValidator) {
				m.On("ValidateAccessToken", "error_token").
					Return(nil, errors.New("validation error"))
			},
			wantStatusCode: http.StatusUnauthorized,
			wantUserID:     "",
			wantNextCalled: false,
		},
		{
			name: "cookie takes precedence over header",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "cookie_token",
				})
				req.Header.Set("Authorization", "Bearer header_token")
			},
			mockSetup: func(m *mockTokenValidator) {
				// Should validate cookie token, not header token
				m.On("ValidateAccessToken", "cookie_token").Return(
					&dto.Claims{UserID: "cookie_user"}, //nolint:exhaustruct // test struct
					nil,
				)
			},
			wantStatusCode: http.StatusOK,
			wantUserID:     "cookie_user",
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := new(mockTokenValidator)
			if tt.mockSetup != nil {
				tt.mockSetup(mockValidator)
			}

			// Create router with middleware
			router := gin.New()
			router.Use(httphandler.AuthMiddleware(mockValidator))

			nextCalled := false
			router.GET("/test", func(c *gin.Context) {
				nextCalled = true
				userID := c.GetString("userID")
				c.JSON(http.StatusOK, gin.H{"userID": userID})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			tt.setupRequest(req)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			assert.Equal(t, tt.wantNextCalled, nextCalled)

			if tt.wantUserID != "" {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					assert.Equal(t, tt.wantUserID, response["userID"])
				}
			}

			mockValidator.AssertExpectations(t)
		})
	}
}
