package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/service"
)

// mockRepository is a mock implementation of domain.Repository for testing.
type mockRepository struct {
	mock.Mock
}

var _ domain.Repository = (*mockRepository)(nil)

func (m *mockRepository) Create(ctx context.Context, recipe *domain.Recipe) error {
	args := m.Called(ctx, recipe)
	return args.Error(0)
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*domain.Recipe, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Recipe), args.Error(1)
}

func (m *mockRepository) FindAll(ctx context.Context) ([]domain.Recipe, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Recipe), args.Error(1)
}

func (m *mockRepository) Search(ctx context.Context, query string) ([]domain.Recipe, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Recipe), args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, recipe *domain.Recipe) error {
	args := m.Called(ctx, recipe)
	return args.Error(0)
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestService_Create(t *testing.T) {
	logger := zerolog.New(nil)

	tests := []struct {
		name      string
		recipe    *domain.Recipe
		mockSetup func(*mockRepository)
		wantErr   error
		wantID    bool
	}{
		{
			name: "success",
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				Title:       "Test Recipe",
				Description: "This is a valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
			},
			mockSetup: func(m *mockRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Recipe")).
					Return(nil)
			},
			wantErr: nil,
			wantID:  true,
		},
		{
			name: "validation error - empty title",
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				Title:       "",
				Description: "Valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
			},
			mockSetup: func(_ *mockRepository) {
				// Repository should not be called
			},
			wantErr: domain.ErrEmptyTitle,
			wantID:  false,
		},
		{
			name: "repository error",
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				Title:       "Test Recipe",
				Description: "This is a valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
			},
			mockSetup: func(m *mockRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Recipe")).
					Return(errors.New("database error"))
			},
			wantErr: errors.New("database error"),
			wantID:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			svc := service.NewService(mockRepo, &logger)
			created, err := svc.Create(context.Background(), tt.recipe)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, created)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, created)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	logger := zerolog.New(nil)
	recipeID := primitive.NewObjectID()

	tests := []struct {
		name       string
		id         string
		mockSetup  func(*mockRepository)
		wantErr    error
		wantRecipe bool
	}{
		{
			name: "success",
			id:   recipeID.Hex(),
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(
						&domain.Recipe{ //nolint:exhaustruct // test struct
							ID:    recipeID,
							Title: "Test",
						},
						nil,
					)
			},
			wantErr:    nil,
			wantRecipe: true,
		},
		{
			name: "not found",
			id:   recipeID.Hex(),
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(nil, domain.ErrNotFound)
			},
			wantErr:    domain.ErrNotFound,
			wantRecipe: false,
		},
		{
			name: "repository error",
			id:   recipeID.Hex(),
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(nil, errors.New("database error"))
			},
			wantErr:    errors.New("database error"),
			wantRecipe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			svc := service.NewService(mockRepo, &logger)
			recipe, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrNotFound) {
					assert.ErrorIs(t, err, domain.ErrNotFound)
				}
				assert.Nil(t, recipe)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, recipe)
				if tt.wantRecipe {
					assert.Equal(t, recipeID, recipe.ID)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetAll(t *testing.T) {
	logger := zerolog.New(nil)

	tests := []struct {
		name      string
		mockSetup func(*mockRepository)
		wantErr   error
		wantCount int
	}{
		{
			name: "success with recipes",
			mockSetup: func(m *mockRepository) {
				m.On("FindAll", mock.Anything).
					Return([]domain.Recipe{
						{Title: "Recipe 1"}, //nolint:exhaustruct // test struct
						{Title: "Recipe 2"}, //nolint:exhaustruct // test struct
					}, nil)
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name: "success empty",
			mockSetup: func(m *mockRepository) {
				m.On("FindAll", mock.Anything).
					Return([]domain.Recipe{}, nil)
			},
			wantErr:   nil,
			wantCount: 0,
		},
		{
			name: "repository error",
			mockSetup: func(m *mockRepository) {
				m.On("FindAll", mock.Anything).
					Return(nil, errors.New("database error"))
			},
			wantErr:   errors.New("database error"),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			svc := service.NewService(mockRepo, &logger)
			recipes, err := svc.GetAll(context.Background())

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, recipes)
			} else {
				require.NoError(t, err)
				assert.Len(t, recipes, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Search(t *testing.T) {
	logger := zerolog.New(nil)

	tests := []struct {
		name      string
		query     string
		mockSetup func(*mockRepository)
		wantErr   error
		wantCount int
	}{
		{
			name:  "search with query",
			query: "pasta",
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "pasta").
					Return([]domain.Recipe{
						{Title: "Pasta Carbonara"}, //nolint:exhaustruct // test struct
						{Title: "Pasta Bolognese"}, //nolint:exhaustruct // test struct
					}, nil)
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:  "empty query - calls GetAll",
			query: "",
			mockSetup: func(m *mockRepository) {
				m.On("FindAll", mock.Anything).
					Return([]domain.Recipe{
						{Title: "Recipe 1"}, //nolint:exhaustruct // test struct
						{Title: "Recipe 2"}, //nolint:exhaustruct // test struct
						{Title: "Recipe 3"}, //nolint:exhaustruct // test struct
					}, nil)
			},
			wantErr:   nil,
			wantCount: 3,
		},
		{
			name:  "search error",
			query: "pasta",
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "pasta").
					Return(nil, errors.New("search error"))
			},
			wantErr:   errors.New("search error"),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			svc := service.NewService(mockRepo, &logger)
			recipes, err := svc.Search(context.Background(), tt.query)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, recipes)
			} else {
				require.NoError(t, err)
				assert.Len(t, recipes, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	logger := zerolog.New(nil)
	recipeID := primitive.NewObjectID()

	tests := []struct {
		name      string
		id        string
		recipe    *domain.Recipe
		mockSetup func(*mockRepository)
		wantErr   error
	}{
		{
			name: "success",
			id:   recipeID.Hex(),
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				Title:       "Updated Recipe",
				Description: "This is a valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 15,
				Portions:    4,
				Tags:        []string{"updated"},
			},
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(
						&domain.Recipe{ //nolint:exhaustruct // test struct
							ID:    recipeID,
							Title: "Old",
						},
						nil,
					)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Recipe")).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   recipeID.Hex(),
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				Title:       "Updated Recipe",
				Description: "This is a valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 15,
				Portions:    4,
				Tags:        []string{"updated"},
			},
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "validation error",
			id:   recipeID.Hex(),
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				Title:       "", // Invalid - empty title
				Description: "Valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 15,
				Portions:    4,
				Tags:        []string{"updated"},
			},
			mockSetup: func(m *mockRepository) {
				// FindByID is called first
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(
						&domain.Recipe{ //nolint:exhaustruct // test struct
							ID:    recipeID,
							Title: "Old",
						},
						nil,
					)
			},
			wantErr: domain.ErrEmptyTitle,
		},
		{
			name: "repository update error",
			id:   recipeID.Hex(),
			recipe: &domain.Recipe{ //nolint:exhaustruct // test struct
				Title:       "Updated Recipe",
				Description: "This is a valid description",
				Ingredients: []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []domain.Step{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 15,
				Portions:    4,
				Tags:        []string{"updated"},
			},
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(
						&domain.Recipe{ //nolint:exhaustruct // test struct
							ID:    recipeID,
							Title: "Old",
						},
						nil,
					)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Recipe")).
					Return(errors.New("update failed"))
			},
			wantErr: errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			svc := service.NewService(mockRepo, &logger)
			updated, err := svc.Update(context.Background(), tt.id, tt.recipe)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrNotFound) ||
					errors.Is(tt.wantErr, domain.ErrValidation) {
					require.ErrorIs(t, err, tt.wantErr)
				}
				assert.Nil(t, updated)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, updated)
				assert.Equal(t, recipeID, updated.ID) // ID should be preserved
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	logger := zerolog.New(nil)
	recipeID := primitive.NewObjectID()

	tests := []struct {
		name      string
		id        string
		mockSetup func(*mockRepository)
		wantErr   error
	}{
		{
			name: "success",
			id:   recipeID.Hex(),
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(
						&domain.Recipe{ //nolint:exhaustruct // test struct
							ID:    recipeID,
							Title: "Test",
						},
						nil)
				m.On("Delete", mock.Anything, recipeID.Hex()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   recipeID.Hex(),
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "delete error",
			id:   recipeID.Hex(),
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID.Hex()).
					Return(
						&domain.Recipe{ //nolint:exhaustruct // test struct
							ID:    recipeID,
							Title: "Test",
						},
						nil,
					)
				m.On("Delete", mock.Anything, recipeID.Hex()).
					Return(errors.New("delete error"))
			},
			wantErr: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			svc := service.NewService(mockRepo, &logger)
			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrNotFound) {
					assert.ErrorIs(t, err, domain.ErrNotFound)
				}
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
