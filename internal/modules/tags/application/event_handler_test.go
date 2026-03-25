package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/infra/messagebus"
	"recipes-desk/internal/modules/tags/application"
	"recipes-desk/internal/modules/tags/application/dto"
	"recipes-desk/internal/modules/tags/application/ports/in"
	"recipes-desk/internal/modules/tags/domain/events"
	"recipes-desk/pkg/pagination"
)

type mockTagService struct {
	mock.Mock
}

var _ in.TagService = (*mockTagService)(nil)

func (m *mockTagService) SearchTags(
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

func (m *mockTagService) GetTagByID(ctx context.Context, id string) (*dto.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Tag), args.Error(1)
}

func (m *mockTagService) CreateTag(ctx context.Context, name string) (*dto.Tag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Tag), args.Error(1)
}

func (m *mockTagService) EnsureTagExists(ctx context.Context, tagName string) error {
	args := m.Called(ctx, tagName)
	return args.Error(0)
}

func (m *mockTagService) EnsureTagsExist(ctx context.Context, tagNames []string) error {
	args := m.Called(ctx, tagNames)
	return args.Error(0)
}

func TestEventHandler_HandleRecipeCreated(t *testing.T) {
	logger := zerolog.New(nil)

	tests := []struct {
		name      string
		payload   events.RecipeCreated
		mockSetup func(*mockTagService)
		wantErr   error
	}{
		{
			name: "success",
			payload: events.RecipeCreated{
				RecipeID: "recipe123",
				Title:    "Pasta",
				Tags:     []string{"Italian", "Dinner"},
			},
			mockSetup: func(m *mockTagService) {
				m.On("EnsureTagsExist", mock.Anything, []string{"Italian", "Dinner"}).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "service error",
			payload: events.RecipeCreated{
				RecipeID: "recipe123",
				Title:    "Pasta",
				Tags:     []string{"Italian"},
			},
			mockSetup: func(m *mockTagService) {
				m.On("EnsureTagsExist", mock.Anything, []string{"Italian"}).
					Return(application.ErrInternal)
			},
			wantErr: application.ErrInternal,
		},
		{
			name: "empty tags",
			payload: events.RecipeCreated{
				RecipeID: "recipe123",
				Title:    "Pasta",
				Tags:     []string{},
			},
			mockSetup: func(m *mockTagService) {
				m.On("EnsureTagsExist", mock.Anything, []string{}).
					Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(mockTagService)
			tt.mockSetup(mockSvc)

			handler := application.NewEventHandler(mockSvc, &logger)

			payload, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			//nolint:exhaustruct // only required fields for test
			msg := messagebus.Message{
				Type:    "recipes.created",
				Payload: payload,
			}

			err = handler.HandleRecipeCreated(msg)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestEventHandler_HandleRecipeCreated_InvalidPayload(t *testing.T) {
	logger := zerolog.New(nil)
	mockSvc := new(mockTagService)
	handler := application.NewEventHandler(mockSvc, &logger)

	//nolint:exhaustruct // only required fields for test
	msg := messagebus.Message{
		Type:    "recipes.created",
		Payload: json.RawMessage(`{invalid json`),
	}

	err := handler.HandleRecipeCreated(msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal event")
}

func TestEventHandler_HandleRecipeDeleted(t *testing.T) {
	logger := zerolog.New(nil)
	mockSvc := new(mockTagService)
	handler := application.NewEventHandler(mockSvc, &logger)

	payload := events.RecipeDeleted{
		RecipeID: "recipe123",
		Tags:     []string{"Italian"},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	//nolint:exhaustruct // only required fields for test
	msg := messagebus.Message{
		Type:    "recipes.deleted",
		Payload: body,
	}

	err = handler.HandleRecipeDeleted(msg)
	require.NoError(t, err)
}

func TestEventHandler_HandleRecipeDeleted_InvalidPayload(t *testing.T) {
	logger := zerolog.New(nil)
	mockSvc := new(mockTagService)
	handler := application.NewEventHandler(mockSvc, &logger)

	//nolint:exhaustruct // only required fields for test
	msg := messagebus.Message{
		Type:    "recipes.deleted",
		Payload: json.RawMessage(`{invalid`),
	}

	err := handler.HandleRecipeDeleted(msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal event")
}
