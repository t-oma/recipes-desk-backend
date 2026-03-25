package httphandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	httphandler "recipes-desk/internal/modules/tags/adapter/in/http"
	"recipes-desk/internal/modules/tags/application"
	"recipes-desk/internal/modules/tags/application/dto"
	"recipes-desk/internal/modules/tags/application/ports/in"
	"recipes-desk/internal/modules/tags/domain"
	"recipes-desk/pkg/pagination"
)

type mockService struct {
	mock.Mock
}

var _ in.TagService = (*mockService)(nil)

func (m *mockService) SearchTags(
	ctx context.Context,
	query string,
	pagn pagination.Request,
) (*pagination.Result[dto.Tag], error) {
	args := m.Called(ctx, query, pagn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pagination.Result[dto.Tag]), args.Error(1)
}

func (m *mockService) GetTagByID(ctx context.Context, id string) (*dto.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Tag), args.Error(1)
}

func (m *mockService) CreateTag(ctx context.Context, name string) (*dto.Tag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Tag), args.Error(1)
}

func (m *mockService) EnsureTagExists(ctx context.Context, tagName string) error {
	args := m.Called(ctx, tagName)
	return args.Error(0)
}

func (m *mockService) EnsureTagsExist(ctx context.Context, tagNames []string) error {
	args := m.Called(ctx, tagNames)
	return args.Error(0)
}

func setupTest() (*gin.Engine, *mockService, *httphandler.Handler) {
	gin.SetMode(gin.TestMode)
	logger := zerolog.New(nil)

	mockSvc := new(mockService)
	h := httphandler.NewHandler(mockSvc, &logger)

	router := gin.New()

	return router, mockSvc, h
}

func TestHandler_GetByID(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tag := &dto.Tag{
		ID:        "id123",
		Name:      "Italian",
		Slug:      "italian",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name           string
		id             string
		mockSetup      func(*mockService)
		wantStatusCode int
	}{
		{
			name: "success",
			id:   "id123",
			mockSetup: func(m *mockService) {
				m.On("GetTagByID", mock.Anything, "id123").Return(tag, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "not found",
			id:   "missing",
			mockSetup: func(m *mockService) {
				m.On("GetTagByID", mock.Anything, "missing").
					Return(nil, domain.ErrNotFound)
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "service error",
			id:   "id123",
			mockSetup: func(m *mockService) {
				m.On("GetTagByID", mock.Anything, "id123").
					Return(nil, application.ErrInternal)
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			tt.mockSetup(mockSvc)

			router.GET("/tags/:id", h.GetByID)

			req := httptest.NewRequest(http.MethodGet, "/tags/"+tt.id, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Search(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	result := &pagination.Result[dto.Tag]{
		Items: []dto.Tag{
			{ID: "id1", Name: "Italian", Slug: "italian", CreatedAt: now, UpdatedAt: now},
		},
		Pagination: pagination.Metadata{Page: 1, Limit: 10, Total: 1, TotalPages: 1},
	}

	tests := []struct {
		name           string
		queryString    string
		mockSetup      func(*mockService)
		wantStatusCode int
		wantItemCount  int
	}{
		{
			name:        "success with query",
			queryString: "?q=italian&page=1&limit=10",
			mockSetup: func(m *mockService) {
				m.On("SearchTags", mock.Anything, "italian",
					pagination.Request{Page: 1, Limit: 10}).
					Return(result, nil)
			},
			wantStatusCode: http.StatusOK,
			wantItemCount:  1,
		},
		{
			name:        "success empty query",
			queryString: "",
			mockSetup: func(m *mockService) {
				m.On("SearchTags", mock.Anything, "",
					pagination.Request{Page: 1, Limit: 20}).
					Return(result, nil)
			},
			wantStatusCode: http.StatusOK,
			wantItemCount:  1,
		},
		{
			name:        "service error",
			queryString: "?q=test",
			mockSetup: func(m *mockService) {
				m.On("SearchTags", mock.Anything, "test",
					pagination.Request{Page: 1, Limit: 20}).
					Return(nil, application.ErrInternal)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantItemCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			tt.mockSetup(mockSvc)

			router.GET("/tags", h.Search)

			req := httptest.NewRequest(http.MethodGet, "/tags"+tt.queryString, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)

			if tt.wantStatusCode == http.StatusOK {
				var resp struct {
					Items []dto.Tag `json:"items"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Len(t, resp.Items, tt.wantItemCount)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Create(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tag := &dto.Tag{
		ID:        "id123",
		Name:      "Italian",
		Slug:      "italian",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name           string
		body           map[string]string
		mockSetup      func(*mockService)
		wantStatusCode int
	}{
		{
			name: "success",
			body: map[string]string{"name": "Italian"},
			mockSetup: func(m *mockService) {
				m.On("CreateTag", mock.Anything, "Italian").Return(tag, nil)
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "invalid json",
			body:           nil,
			mockSetup:      func(_ *mockService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "validation error - empty name",
			body:           map[string]string{"name": ""},
			mockSetup:      func(_ *mockService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "conflict",
			body: map[string]string{"name": "Italian"},
			mockSetup: func(m *mockService) {
				m.On("CreateTag", mock.Anything, "Italian").
					Return(nil, domain.ErrConflict)
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name: "service error",
			body: map[string]string{"name": "Italian"},
			mockSetup: func(m *mockService) {
				m.On("CreateTag", mock.Anything, "Italian").
					Return(nil, application.ErrInternal)
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc, h := setupTest()
			tt.mockSetup(mockSvc)

			router.POST("/tags", h.Create)

			var body *bytes.Buffer
			if tt.body != nil {
				b, _ := json.Marshal(tt.body)
				body = bytes.NewBuffer(b)
			} else {
				body = bytes.NewBufferString("{invalid")
			}

			req := httptest.NewRequest(http.MethodPost, "/tags", body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
