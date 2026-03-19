//go:build integration
// +build integration

package rabbitmq_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/rabbitmq"
	"recipes-desk/pkg/pool"
	"recipes-desk/pkg/testutils"
)

// MockEvent is a mock event for testing.
type MockEvent struct {
	mock.Mock
}

func (m *MockEvent) Type() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEvent) RoutingKey() string {
	args := m.Called()
	return args.String(0)
}

func TestIntegration_Publisher_NewPublisher(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	log := zerolog.New(zerolog.NewConsoleWriter())
	connManager := setupRabbitMQFromContainer(t, container, &log)
	defer connManager.Stop()

	chanpool, err := rabbitmq.NewChannelPool(context.Background(), connManager, 5, &log)
	require.NoError(t, err)
	defer chanpool.Close(context.Background())

	tests := []struct {
		name    string
		config  rabbitmq.PublisherConfig
		setup   func(pool.Pool[*amqp.Channel])
		wantErr error
	}{
		{
			name: "success",
			config: rabbitmq.PublisherConfig{
				Exchange:     "test.exchange",
				ExchangeType: "topic",
			},
			setup: func(pool pool.Pool[*amqp.Channel]) {
			},
			wantErr: nil,
		},
		{
			name: "fails - pool get returns error",
			config: rabbitmq.PublisherConfig{
				Exchange:     "test.exchange",
				ExchangeType: "topic",
			},
			setup: func(pool pool.Pool[*amqp.Channel]) {
				pool.Close(context.Background())
			},
			wantErr: rabbitmq.ErrChannelPoolClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(chanpool)
			}

			publisher, err := rabbitmq.NewPublisher(context.Background(), tt.config, chanpool, &log)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, publisher)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, publisher)
			}
		})
	}
}

func TestPublisher_Publish(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	log := zerolog.New(zerolog.NewConsoleWriter())
	connManager := setupRabbitMQFromContainer(t, container, &log)
	defer connManager.Stop()

	pool, err := rabbitmq.NewChannelPool(context.Background(), connManager, 5, &log)
	require.NoError(t, err)
	defer pool.Close(context.Background())

	tests := []struct {
		name      string
		config    rabbitmq.PublisherConfig
		setupMock func(*MockEvent)
		wantErr   bool
	}{
		{
			name: "successfully publishes event",
			config: rabbitmq.PublisherConfig{
				Exchange:     "test.exchange",
				ExchangeType: "topic",
			},
			setupMock: func(e *MockEvent) {
				e.On("Type").Return("TestEvent").Once()
				e.On("RoutingKey").Return("test.event").Once()
			},
			wantErr: false,
		},
		{
			name: "fails when pool get returns error",
			config: rabbitmq.PublisherConfig{
				Exchange:     "test.exchange",
				ExchangeType: "topic",
			},
			setupMock: func(_ *MockEvent) {
			},
			wantErr: true,
		},
		{
			name: "fails when channel is closed",
			config: rabbitmq.PublisherConfig{
				Exchange:     "test.exchange",
				ExchangeType: "topic",
			},
			setupMock: func(e *MockEvent) {
				e.On("Type").Return("TestEvent").Once()
				e.On("RoutingKey").Return("test.event").Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEvent := new(MockEvent)
			tt.setupMock(mockEvent)

			ctx := context.Background()

			publisher, err := rabbitmq.NewPublisher(ctx, tt.config, pool, &log)
			require.NoError(t, err)

			err = publisher.Publish(ctx, mockEvent)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockEvent.AssertExpectations(t)
		})
	}
}

func TestPublisher_PublishWithConfirm(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	log := zerolog.New(zerolog.NewConsoleWriter())
	connManager := setupRabbitMQFromContainer(t, container, &log)
	defer connManager.Stop()

	pool, err := rabbitmq.NewChannelPool(context.Background(), connManager, 5, &log)
	require.NoError(t, err)
	defer pool.Close(context.Background())

	tests := []struct {
		name      string
		config    rabbitmq.PublisherConfig
		setupMock func(*MockEvent)
		timeout   time.Duration
		wantErr   bool
	}{
		{
			name: "successfully publishes with confirmation",
			config: rabbitmq.PublisherConfig{
				Exchange:     "test.exchange",
				ExchangeType: "topic",
			},
			setupMock: func(e *MockEvent) {
				e.On("Type").Return("TestEvent").Once()
				e.On("RoutingKey").Return("test.event").Once()
			},
			timeout: 5 * time.Second,
			wantErr: false,
		},
		{
			name: "times out waiting for confirmation",
			config: rabbitmq.PublisherConfig{
				Exchange:     "test.exchange",
				ExchangeType: "topic",
			},
			setupMock: func(e *MockEvent) {
				e.On("Type").Return("TestEvent").Once()
				e.On("RoutingKey").Return("test.event").Once()
			},
			timeout: 1 * time.Millisecond,
			wantErr: true,
		},
		{
			name: "fails when message is nacked",
			config: rabbitmq.PublisherConfig{
				Exchange:     "test.exchange",
				ExchangeType: "topic",
			},
			setupMock: func(e *MockEvent) {
				e.On("Type").Return("TestEvent").Once()
				e.On("RoutingKey").Return("test.event").Once()
			},
			timeout: 5 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEvent := new(MockEvent)
			tt.setupMock(mockEvent)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			publisher, err := rabbitmq.NewPublisher(ctx, tt.config, pool, &log)
			require.NoError(t, err)

			err = publisher.PublishWithConfirm(ctx, mockEvent)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockEvent.AssertExpectations(t)
		})
	}
}
