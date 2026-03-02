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

	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/handler"
)

// mockService is a mock implementation of handler.RecipeService for testing.
type mockService struct {
	mock.Mock
}

func (m *mockService) Create(ctx context.Context, recipe *domain.Recipe) (*domain.Recipe, error) {
	args := m.Called(ctx, recipe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Recipe), args.Error(1)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*domain.Recipe, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Recipe), args.Error(1)
}

func (m *mockService) GetAll(ctx context.Context) ([]domain.Recipe, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Recipe), args.Error(1)
}

func (m *mockService) Search(ctx context.Context, query string) ([]domain.Recipe, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Recipe), args.Error(1)
}

func (m *mockService) Update(
	ctx context.Context,
	id string,
	recipe *domain.Recipe,
) (*domain.Recipe, error) {
	args := m.Called(ctx, id, recipe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Recipe), args.Error(1)
}

func (m *mockService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTest() (*gin.Engine, *mockService, *handler.Handler) {
	gin.SetMode(gin.TestMode)
	logger := zerolog.New(nil)

	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc, &logger)

	router := gin.New()

	return router, mockSvc, h
}

func TestHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*mockService)
		wantStatusCode int
		wantRecipes    int
	}{
		{
			name: "success with recipes",
			mockSetup: func(m *mockService) {
				m.On("GetAll", mock.Anything).Return([]domain.Recipe{
					{ //nolint:exhaustruct // test struct
						ID:    "id123",
						Title: "Recipe 1",
					},
					{ //nolint:exhaustruct // test struct
						ID:    "id456",
						Title: "Recipe 2",
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRecipes:    2,
		},
		{
			name: "success empty",
			mockSetup: func(m *mockService) {
				m.On("GetAll", mock.Anything).Return([]domain.Recipe{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRecipes:    0,
		},
		{
			name: "service error",
			mockSetup: func(m *mockService) {
				m.On("GetAll", mock.Anything).Return(nil, errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRecipes:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			tt.mockSetup(mockSvc)

			router.GET("/recipes", h.List)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/recipes", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantStatusCode == http.StatusOK {
				var response handler.ListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response.Recipes, tt.wantRecipes)
				assert.Equal(t, tt.wantRecipes, response.Count)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetByID(t *testing.T) {
	recipeID := "abc123"

	tests := []struct {
		name           string
		id             string
		mockSetup      func(*mockService)
		wantStatusCode int
		wantRecipe     bool
	}{
		{
			name: "success",
			id:   recipeID,
			mockSetup: func(m *mockService) {
				m.On("GetByID", mock.Anything, recipeID).
					Return(&domain.Recipe{
						ID:          recipeID,
						Title:       "Test Recipe",
						Description: "Test Description",
						Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
						Steps:       []domain.Step{{Order: 1, Description: "Step 1", Duration: 10}},
						CookingTime: 30,
						Portions:    4,
						Tags:        []string{"test"},
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRecipe:     true,
		},
		{
			name: "not found",
			id:   recipeID,
			mockSetup: func(m *mockService) {
				m.On("GetByID", mock.Anything, recipeID).
					Return(nil, domain.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
			wantRecipe:     false,
		},
		{
			name: "service error",
			id:   recipeID,
			mockSetup: func(m *mockService) {
				m.On("GetByID", mock.Anything, recipeID).
					Return(nil, errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRecipe:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			tt.mockSetup(mockSvc)

			router.GET("/recipes/:id", h.GetByID)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/recipes/"+tt.id, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantRecipe {
				var response handler.RecipeResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, recipeID, response.ID)
				assert.Equal(t, "Test Recipe", response.Title)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockSetup      func(*mockService)
		wantStatusCode int
		wantRecipes    int
	}{
		{
			name:  "success",
			query: "pasta",
			mockSetup: func(m *mockService) {
				m.On("Search", mock.Anything, "pasta").
					Return([]domain.Recipe{
						{Title: "Pasta Carbonara"}, //nolint:exhaustruct // test struct
						{Title: "Pasta Bolognese"}, //nolint:exhaustruct // test struct
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRecipes:    2,
		},
		{
			name:  "empty query",
			query: "",
			mockSetup: func(m *mockService) {
				m.On("Search", mock.Anything, "").
					Return([]domain.Recipe{}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRecipes:    0,
		},
		{
			name:  "service error",
			query: "pasta",
			mockSetup: func(m *mockService) {
				m.On("Search", mock.Anything, "pasta").
					Return(nil, errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRecipes:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			tt.mockSetup(mockSvc)

			router.GET("/recipes/search", h.Search)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/recipes/search?q="+tt.query, nil)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantStatusCode == http.StatusOK {
				var response handler.ListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response.Recipes, tt.wantRecipes)
				assert.Equal(t, tt.wantRecipes, response.Count)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Create(t *testing.T) {
	recipeID := "abc123"

	tests := []struct {
		name           string
		userID         string
		body           map[string]any
		mockSetup      func(*mockService)
		wantStatusCode int
		wantCreated    bool
	}{
		{
			name:   "success",
			userID: "user-id",
			body: map[string]any{
				"title":       "New Recipe",
				"description": "This is a valid description that is long enough",
				"ingredients": []map[string]any{
					{"name": "Ingredient 1", "amount": 100, "unit": "g"},
				},
				"steps": []map[string]any{
					{"order": 1, "description": "Step 1", "duration": 10},
				},
				"cookingTime": 30,
				"portions":    4,
				"tags":        []string{"italian", "pasta"},
			},
			mockSetup: func(m *mockService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Recipe")).
					Return(&domain.Recipe{ //nolint:exhaustruct // test struct
						ID:    recipeID,
						Title: "New Recipe",
					}, nil)
			},
			wantStatusCode: http.StatusCreated,
			wantCreated:    true,
		},
		{
			name:   "validation error",
			userID: "user-id",
			body: map[string]any{
				"title":       "AB", // Title too short (less than 3 characters)
				"description": "Valid description",
				"ingredients": []map[string]any{
					{"name": "Ingredient 1", "amount": 100, "unit": "g"},
				},
				"steps": []map[string]any{
					{"order": 1, "description": "Step 1", "duration": 10},
				},
				"cookingTime": 30,
				"portions":    4,
				"tags":        []string{"test"},
			},
			mockSetup: func(m *mockService) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(nil, domain.ErrInvalidTitleLength)
			},
			wantStatusCode: http.StatusBadRequest,
			wantCreated:    false,
		},
		{
			name:   "invalid json",
			userID: "user-id",
			body:   nil, // Will send invalid JSON
			mockSetup: func(_ *mockService) {
				// Service should not be called
			},
			wantStatusCode: http.StatusBadRequest,
			wantCreated:    false,
		},
		{
			name:   "unauthorized",
			userID: "",
			body:   nil,
			mockSetup: func(_ *mockService) {
				// Service should not be called
			},
			wantStatusCode: http.StatusUnauthorized,
			wantCreated:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			router.POST("/recipes", func(c *gin.Context) {
				if tt.userID != "" {
					c.Set("userID", tt.userID)
				}
				h.Create(c)
			})

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			} else {
				body = []byte(`{invalid json`)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/recipes", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantCreated {
				var response handler.RecipeResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, recipeID, response.ID)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Update(t *testing.T) {
	recipeID := "abc123"

	tests := []struct {
		name           string
		id             string
		body           map[string]any
		mockSetup      func(*mockService)
		wantStatusCode int
	}{
		{
			name: "success",
			id:   recipeID,
			body: map[string]any{
				"title":       "Updated Recipe",
				"description": "This is a valid description that is long enough",
				"ingredients": []map[string]any{
					{"name": "Ingredient 1", "amount": 100, "unit": "g"},
				},
				"steps": []map[string]any{
					{"order": 1, "description": "Step 1", "duration": 10},
				},
				"cookingTime": 30,
				"portions":    4,
				"tags":        []string{"updated"},
			},
			mockSetup: func(m *mockService) {
				m.On("Update", mock.Anything, recipeID, mock.AnythingOfType("*domain.Recipe")).
					Return(
						&domain.Recipe{ //nolint:exhaustruct // test struct
							ID:    "id123",
							Title: "Updated Recipe",
						},
						nil,
					)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "not found",
			id:   recipeID,
			body: map[string]any{
				"title":       "Updated Recipe",
				"description": "This is a valid description that is long enough",
				"ingredients": []map[string]any{
					{"name": "Ingredient 1", "amount": 100, "unit": "g"},
				},
				"steps": []map[string]any{
					{"order": 1, "description": "Step 1", "duration": 10},
				},
				"cookingTime": 30,
				"portions":    4,
				"tags":        []string{"updated"},
			},
			mockSetup: func(m *mockService) {
				m.On("Update", mock.Anything, recipeID, mock.AnythingOfType("*domain.Recipe")).
					Return(nil, domain.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "http validation error",
			id:   recipeID,
			body: map[string]any{
				"title":       "", // empty title
				"description": "This is a valid description that is long enough",
				"ingredients": []map[string]any{
					{"name": "Ingredient 1", "amount": 100, "unit": "g"},
				},
				"steps": []map[string]any{
					{"order": 1, "description": "Step 1", "duration": 10},
				},
				"cookingTime": 30,
				"portions":    4,
				"tags":        []string{"updated"},
			},
			mockSetup: func(_ *mockService) {
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "validation error",
			id:   recipeID,
			body: map[string]any{
				"title":       "AB", // Title too short
				"description": "Valid description",
				"ingredients": []map[string]any{
					{"name": "Ingredient 1", "amount": 100, "unit": "g"},
				},
				"steps": []map[string]any{
					{"order": 1, "description": "Step 1", "duration": 10},
				},
				"cookingTime": 30,
				"portions":    4,
				"tags":        []string{"updated"},
			},
			mockSetup: func(m *mockService) {
				m.On("Update", mock.Anything, recipeID, mock.Anything).
					Return(nil, domain.ErrInvalidTitleLength)
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			tt.mockSetup(mockSvc)

			router.PUT("/recipes/:id", h.Update)

			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/recipes/"+tt.id, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	recipeID := "abc123"

	tests := []struct {
		name           string
		id             string
		mockSetup      func(*mockService)
		wantStatusCode int
	}{
		{
			name: "success",
			id:   recipeID,
			mockSetup: func(m *mockService) {
				m.On("Delete", mock.Anything, recipeID).Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "not found",
			id:   recipeID,
			mockSetup: func(m *mockService) {
				m.On("Delete", mock.Anything, recipeID).Return(domain.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "service error",
			id:   recipeID,
			mockSetup: func(m *mockService) {
				m.On("Delete", mock.Anything, recipeID).Return(errors.New("database error"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			tt.mockSetup(mockSvc)

			router.DELETE("/recipes/:id", h.Delete)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/recipes/"+tt.id, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
