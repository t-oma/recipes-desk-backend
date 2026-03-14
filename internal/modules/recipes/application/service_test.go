package application_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/recipes/application"
	"recipes-desk/internal/modules/recipes/application/dto"
	"recipes-desk/internal/modules/recipes/domain"
	"recipes-desk/internal/modules/recipes/domain/entity"
	"recipes-desk/internal/modules/recipes/domain/fixtures"
	"recipes-desk/internal/modules/recipes/domain/ports"
	"recipes-desk/internal/modules/recipes/domain/valueobject"
	"recipes-desk/pkg/pagination"
)

const _recipeTypeString = "*entity.Recipe"

// mockRepository is a mock implementation of domain.Repository for testing.
type mockRepository struct {
	mock.Mock
}

var _ ports.RecipeRepository = (*mockRepository)(nil)

func (m *mockRepository) Create(
	ctx context.Context,
	recipe *entity.Recipe,
) (*entity.Recipe, error) {
	args := m.Called(ctx, recipe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Recipe), args.Error(1)
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*entity.Recipe, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Recipe), args.Error(1)
}

func (m *mockRepository) Search(
	ctx context.Context,
	query string,
	skip, limit int64,
) ([]entity.Recipe, int64, error) {
	args := m.Called(ctx, query, skip, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.Recipe), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) Update(
	ctx context.Context,
	recipe *entity.Recipe,
) (*entity.Recipe, error) {
	args := m.Called(ctx, recipe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Recipe), args.Error(1)
}

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

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestService_Create(t *testing.T) {
	logger := zerolog.New(nil)

	tests := []struct {
		name      string
		input     dto.CreateRecipeInput
		mockSetup func(*mockRepository, *mockIDGenerator)
		wantErr   error
		wantID    bool
	}{
		{
			name: "success",
			input: dto.CreateRecipeInput{
				Title:       "Test Recipe",
				Description: "This is a valid description",
				Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
				UserID:      "user-id",
			},
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")

				m.On("Create", mock.Anything, mock.AnythingOfType(_recipeTypeString)).
					Return(fixtures.NewRecipe(t, "id123", "author123", "Test Recipe"), nil)
			},
			wantErr: nil,
			wantID:  true,
		},
		{
			name: "validation error - empty title",
			input: dto.CreateRecipeInput{
				Title:       "",
				Description: "Valid description",
				Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
				UserID:      "user-id",
			},
			mockSetup: func(_ *mockRepository, _ *mockIDGenerator) {
				// Repository should not be called
			},
			wantErr: valueobject.ErrTitleEmpty,
			wantID:  false,
		},
		{
			name: "validation error - empty ingridient name",
			input: dto.CreateRecipeInput{
				Title:       "Test Recipe",
				Description: "Valid description",
				Ingredients: []dto.IngredientInput{{Name: "", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
				UserID:      "user-id",
			},
			mockSetup: func(_ *mockRepository, _ *mockIDGenerator) {
				// Repository should not be called
			},
			wantErr: valueobject.ErrIngredientEmptyName,
			wantID:  false,
		},
		{
			name: "validation error - negative step duration",
			input: dto.CreateRecipeInput{
				Title:       "Test Recipe",
				Description: "Valid description",
				Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: -1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
				UserID:      "user-id",
			},
			mockSetup: func(_ *mockRepository, _ *mockIDGenerator) {
				// Repository should not be called
			},
			wantErr: valueobject.ErrStepNegativeDuration,
			wantID:  false,
		},
		{
			name: "validation error - empty tag name",
			input: dto.CreateRecipeInput{
				Title:       "Test Recipe",
				Description: "Valid description",
				Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{""},
				UserID:      "user-id",
			},
			mockSetup: func(_ *mockRepository, _ *mockIDGenerator) {
				// Repository should not be called
			},
			wantErr: valueobject.ErrTagEmptyName,
			wantID:  false,
		},
		{
			name: "repository error",
			input: dto.CreateRecipeInput{
				Title:       "Test Recipe",
				Description: "This is a valid description",
				Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 10,
				Portions:    2,
				Tags:        []string{"test"},
				UserID:      "user-id",
			},
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")

				m.On("Create", mock.Anything, mock.AnythingOfType(_recipeTypeString)).
					Return(nil, domain.ErrDatabase)
			},
			wantErr: application.ErrInternal,
			wantID:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			mockIDGen := new(mockIDGenerator)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockIDGen)
			}

			svc := application.NewService(mockRepo, mockIDGen, &logger, 10, 10)
			created, err := svc.Create(context.Background(), tt.input)

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
	recipe := fixtures.NewRecipe(t, "id123", "author123", "Test Recipe")
	recipeID := recipe.ID().String()

	tests := []struct {
		name       string
		id         string
		mockSetup  func(*mockRepository)
		wantErr    error
		wantRecipe bool
	}{
		{
			name: "success",
			id:   recipeID,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(recipe, nil)
			},
			wantErr:    nil,
			wantRecipe: true,
		},
		{
			name: "not found",
			id:   recipeID,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(nil, domain.ErrNotFound)
			},
			wantErr:    domain.ErrNotFound,
			wantRecipe: false,
		},
		{
			name: "repository error",
			id:   recipeID,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(nil, domain.ErrDatabase)
			},
			wantErr:    application.ErrInternal,
			wantRecipe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			idGen := new(mockIDGenerator)
			svc := application.NewService(mockRepo, idGen, &logger, 10, 10)
			recipe, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
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

func TestService_Search(t *testing.T) {
	logger := zerolog.New(nil)

	recipes := []entity.Recipe{
		*fixtures.NewRecipe(t, "id123", "author123", "Pasta Carbonara"),
		*fixtures.NewRecipe(t, "id456", "author123", "Pasta Bolognese"),
		*fixtures.NewRecipe(t, "id789", "author123", "Chicken Curry"),
	}
	recipesCount := len(recipes)

	tests := []struct {
		name      string
		query     string
		pagn      pagination.Request
		mockSetup func(*mockRepository)
		wantErr   error
		wantCount int
		wantPage  int
		wantLimit int
	}{
		{
			name:  "search with query",
			query: "pasta",
			pagn:  pagination.Request{Page: 1, Limit: 10},
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "pasta", pagination.Skip(1, 10), int64(10)).
					Return(recipes[0:2], int64(2), nil)
			},
			wantErr:   nil,
			wantCount: 2,
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:  "empty query",
			query: "",
			pagn:  pagination.Request{Page: 1, Limit: 10},
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "", pagination.Skip(1, 10), int64(10)).
					Return(recipes, int64(recipesCount), nil)
			},
			wantErr:   nil,
			wantCount: recipesCount,
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:  "search error",
			query: "pasta",
			pagn:  pagination.Request{Page: 1, Limit: 10},
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "pasta", pagination.Skip(1, 10), int64(10)).
					Return(nil, int64(0), domain.ErrDatabase)
			},
			wantErr:   application.ErrInternal,
			wantCount: 0,
			wantPage:  1,
			wantLimit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			idGen := new(mockIDGenerator)
			svc := application.NewService(mockRepo, idGen, &logger, 10, 10)
			result, err := svc.Search(
				context.Background(),
				tt.query,
				tt.pagn,
			)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result.Items, tt.wantCount)
				assert.Equal(t, tt.wantPage, result.Pagination.Page)
				assert.Equal(t, tt.wantLimit, result.Pagination.Limit)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	logger := zerolog.New(nil)
	recipe := fixtures.NewRecipe(t, "id123", "author123", "Original Title")
	recipeID := recipe.ID().String()
	authorID := recipe.AuthorID().String()

	validInput := dto.UpdateRecipeInput{
		Title:       "Updated Recipe",
		Description: "This is a valid description",
		Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
		Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
		CookingTime: 15,
		Portions:    4,
		Tags:        []string{"updated"},
	}

	tests := []struct {
		name      string
		userID    string
		id        string
		input     dto.UpdateRecipeInput
		mockSetup func(*mockRepository)
		wantErr   error
	}{
		{
			name:   "success",
			userID: authorID,
			id:     recipeID,
			input:  validInput,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(recipe, nil)

				newTitle, err := valueobject.NewTitle("Updated Title")
				require.NoError(t, err)
				recipe.UpdateTitle(newTitle)

				m.On("Update", mock.Anything, mock.AnythingOfType(_recipeTypeString)).
					Return(recipe, nil)
			},
			wantErr: nil,
		},
		{
			name:   "not found",
			userID: authorID,
			id:     recipeID,
			input:  validInput,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name:   "validation error - empty title",
			userID: authorID,
			id:     recipeID,
			input: dto.UpdateRecipeInput{
				Title:       "", // Invalid - empty title
				Description: "Valid description",
				Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 15,
				Portions:    4,
				Tags:        []string{"updated"},
			},
			mockSetup: func(_ *mockRepository) {
			},
			wantErr: valueobject.ErrTitleEmpty,
		},
		{
			name:   "validation error - empty ingridient name",
			userID: authorID,
			id:     recipeID,
			input: dto.UpdateRecipeInput{
				Title:       "Test Recipe",
				Description: "Valid description",
				Ingredients: []dto.IngredientInput{{Name: "", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 15,
				Portions:    4,
				Tags:        []string{"updated"},
			},
			mockSetup: func(_ *mockRepository) {
			},
			wantErr: valueobject.ErrIngredientEmptyName,
		},
		{
			name:   "validation error - empty tag name",
			userID: authorID,
			id:     recipeID,
			input: dto.UpdateRecipeInput{
				Title:       "Test", // Invalid - empty title
				Description: "Valid description",
				Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: 1}},
				CookingTime: 15,
				Portions:    4,
				Tags:        []string{""},
			},
			mockSetup: func(_ *mockRepository) {
			},
			wantErr: valueobject.ErrTagEmptyName,
		},
		{
			name:   "validation error - negative step duration",
			userID: authorID,
			id:     recipeID,
			input: dto.UpdateRecipeInput{
				Title:       "Test", // Invalid - empty title
				Description: "Valid description",
				Ingredients: []dto.IngredientInput{{Name: "Test", Amount: 1, Unit: "g"}},
				Steps:       []dto.StepInput{{Order: 1, Description: "Step", Duration: -1}},
				CookingTime: 15,
				Portions:    4,
				Tags:        []string{"testtag"},
			},
			mockSetup: func(_ *mockRepository) {
			},
			wantErr: valueobject.ErrStepNegativeDuration,
		},
		{
			name:   "repository update error",
			userID: authorID,
			id:     recipeID,
			input:  validInput,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(recipe, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType(_recipeTypeString)).
					Return(nil, domain.ErrDatabase)
			},
			wantErr: application.ErrInternal,
		},
		{
			name:   "forbidden",
			userID: "not-author",
			id:     recipeID,
			input:  validInput,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(recipe, nil)
			},
			wantErr: domain.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			idGen := new(mockIDGenerator)
			svc := application.NewService(mockRepo, idGen, &logger, 10, 10)
			updated, err := svc.Update(context.Background(), tt.userID, tt.id, tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
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
	recipe := fixtures.NewRecipe(t, "id123", "author123", "Test Recipe")
	recipeID := recipe.ID().String()
	authorID := recipe.AuthorID().String()

	tests := []struct {
		name      string
		userID    string
		id        string
		mockSetup func(*mockRepository)
		wantErr   error
	}{
		{
			name:   "success",
			id:     recipeID,
			userID: authorID,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(recipe, nil)
				m.On("Delete", mock.Anything, recipeID).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "not found",
			id:     recipeID,
			userID: authorID,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name:   "delete error",
			id:     recipeID,
			userID: authorID,
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(recipe, nil)
				m.On("Delete", mock.Anything, recipeID).
					Return(domain.ErrDatabase)
			},
			wantErr: application.ErrInternal,
		},
		{
			name:   "forbidden",
			id:     recipeID,
			userID: "not-author",
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, recipeID).
					Return(recipe, nil)
				// Delete should not be called because recipe is not owned by user
			},
			wantErr: application.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			idGen := new(mockIDGenerator)
			svc := application.NewService(mockRepo, idGen, &logger, 10, 10)
			err := svc.Delete(context.Background(), tt.userID, tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
