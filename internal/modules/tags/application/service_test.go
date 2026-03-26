package application_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/modules/tags/application"
	"recipes-desk/internal/modules/tags/domain"
	"recipes-desk/internal/modules/tags/domain/entity"
	"recipes-desk/internal/modules/tags/domain/fixtures"
	"recipes-desk/internal/modules/tags/domain/ports"
	"recipes-desk/internal/modules/tags/domain/valueobject"
	"recipes-desk/pkg/pagination"
)

const _tagTypeString = "*entity.Tag"

type mockRepository struct {
	mock.Mock
}

var _ ports.TagRepository = (*mockRepository)(nil)

func (m *mockRepository) Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	args := m.Called(ctx, tag)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Tag), args.Error(1)
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*entity.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Tag), args.Error(1)
}

func (m *mockRepository) Search(
	ctx context.Context,
	query string,
	skip, limit int64,
) ([]entity.Tag, int64, error) {
	args := m.Called(ctx, query, skip, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.Tag), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) Exists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

type mockIDGenerator struct {
	mock.Mock
}

var _ ports.IDGenerator = (*mockIDGenerator)(nil)

func (m *mockIDGenerator) Generate() string {
	args := m.Called()
	return args.String(0)
}

func newService(repo ports.TagRepository, idGen ports.IDGenerator) *application.Service {
	logger := zerolog.New(nil)
	return application.NewService(repo, idGen, &logger, 100, 20)
}

func TestService_CreateTag(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		mockSetup func(*mockRepository, *mockIDGenerator)
		wantErr   error
	}{
		{
			name:      "success",
			inputName: "Italian",
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(fixtures.NewTag(t, "id123", "Italian"), nil)
			},
			wantErr: nil,
		},
		{
			name:      "validation error - empty name",
			inputName: "",
			mockSetup: func(_ *mockRepository, _ *mockIDGenerator) {},
			wantErr:   valueobject.ErrTagNameEmpty,
		},
		{
			name:      "validation error - too short",
			inputName: "A",
			mockSetup: func(_ *mockRepository, _ *mockIDGenerator) {},
			wantErr:   valueobject.ErrTagNameTooShort,
		},
		{
			name:      "conflict - duplicate slug",
			inputName: "Italian",
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(nil, domain.ErrConflict)
			},
			wantErr: domain.ErrConflict,
		},
		{
			name:      "database error",
			inputName: "Italian",
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(nil, domain.ErrDatabase)
			},
			wantErr: application.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			mockIDGen := new(mockIDGenerator)
			tt.mockSetup(mockRepo, mockIDGen)

			svc := newService(mockRepo, mockIDGen)
			tag, err := svc.CreateTag(context.Background(), tt.inputName)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, tag)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tag)
			}

			mockRepo.AssertExpectations(t)
			mockIDGen.AssertExpectations(t)
		})
	}
}

func TestService_GetTagByID(t *testing.T) {
	tag := fixtures.NewTag(t, "id123", "Italian")

	tests := []struct {
		name      string
		id        string
		mockSetup func(*mockRepository)
		wantErr   error
	}{
		{
			name: "success",
			id:   "id123",
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, "id123").Return(tag, nil)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   "missing",
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, "missing").Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "database error",
			id:   "id123",
			mockSetup: func(m *mockRepository) {
				m.On("FindByID", mock.Anything, "id123").Return(nil, domain.ErrDatabase)
			},
			wantErr: application.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			svc := newService(mockRepo, new(mockIDGenerator))
			tag, err := svc.GetTagByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, tag)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tag)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_SearchTags(t *testing.T) {
	tags := []entity.Tag{
		*fixtures.NewTag(t, "id1", "Italian"),
		*fixtures.NewTag(t, "id2", "Indian"),
	}

	tests := []struct {
		name      string
		query     string
		pagn      pagination.Request
		mockSetup func(*mockRepository)
		wantErr   error
		wantCount int
	}{
		{
			name:  "success",
			query: "ita",
			pagn:  pagination.Request{Page: 1, Limit: 10},
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "ita", int64(0), int64(10)).
					Return(tags, int64(2), nil)
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:  "empty result",
			query: "zzz",
			pagn:  pagination.Request{Page: 1, Limit: 10},
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "zzz", int64(0), int64(10)).
					Return([]entity.Tag{}, int64(0), nil)
			},
			wantErr:   nil,
			wantCount: 0,
		},
		{
			name:  "pagination normalization",
			query: "",
			pagn:  pagination.Request{Page: 0, Limit: 0},
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "", int64(0), int64(20)).
					Return(tags, int64(2), nil)
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:  "database error",
			query: "test",
			pagn:  pagination.Request{Page: 1, Limit: 10},
			mockSetup: func(m *mockRepository) {
				m.On("Search", mock.Anything, "test", int64(0), int64(10)).
					Return(nil, int64(0), domain.ErrDatabase)
			},
			wantErr:   application.ErrInternal,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			tt.mockSetup(mockRepo)

			svc := newService(mockRepo, new(mockIDGenerator))
			result, err := svc.SearchTags(context.Background(), tt.query, tt.pagn)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result.Items, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_EnsureTagExists(t *testing.T) {
	tag := fixtures.NewTag(t, "id123", "Italian")

	tests := []struct {
		name      string
		tagName   string
		mockSetup func(*mockRepository, *mockIDGenerator)
		wantErr   error
	}{
		{
			name:    "creates new tag",
			tagName: "Italian",
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(tag, nil)
			},
			wantErr: nil,
		},
		{
			name:    "tag already exists - no error",
			tagName: "Italian",
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(nil, domain.ErrConflict)
			},
			wantErr: nil,
		},
		{
			name:    "database error propagates",
			tagName: "Italian",
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(nil, domain.ErrDatabase)
			},
			wantErr: application.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			mockIDGen := new(mockIDGenerator)
			tt.mockSetup(mockRepo, mockIDGen)

			svc := newService(mockRepo, mockIDGen)
			err := svc.EnsureTagExists(context.Background(), tt.tagName)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_EnsureTagsExist(t *testing.T) {
	tag := fixtures.NewTag(t, "id123", "Italian")

	tests := []struct {
		name      string
		tagNames  []string
		mockSetup func(*mockRepository, *mockIDGenerator)
		wantErr   error
	}{
		{
			name:     "all tags created",
			tagNames: []string{"Italian", "Spicy"},
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123").Twice()
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(tag, nil).Twice()
			},
			wantErr: nil,
		},
		{
			name:     "all tags already exist",
			tagNames: []string{"Italian", "Spicy"},
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123").Twice()
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(nil, domain.ErrConflict).Twice()
			},
			wantErr: nil,
		},
		{
			name:      "empty list",
			tagNames:  []string{},
			mockSetup: func(_ *mockRepository, _ *mockIDGenerator) {},
			wantErr:   nil,
		},
		{
			name:     "first fails - stops loop",
			tagNames: []string{"Italian", "Spicy"},
			mockSetup: func(m *mockRepository, mID *mockIDGenerator) {
				mID.On("Generate").Return("id123")
				m.On("Create", mock.Anything, mock.AnythingOfType(_tagTypeString)).
					Return(nil, domain.ErrDatabase)
			},
			wantErr: application.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			mockIDGen := new(mockIDGenerator)
			tt.mockSetup(mockRepo, mockIDGen)

			svc := newService(mockRepo, mockIDGen)
			err := svc.EnsureTagsExist(context.Background(), tt.tagNames)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
