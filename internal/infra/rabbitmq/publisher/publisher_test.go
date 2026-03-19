//go:build integration
// +build integration

package publisher_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	amqp "github.com/rabbitmq/amqp091-go"
	rabbitmqtc "github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	"recipes-desk/internal/infra/rabbitmq"
	"recipes-desk/internal/infra/rabbitmq/connection"
	"recipes-desk/internal/infra/rabbitmq/publisher"
	"recipes-desk/pkg/pool"
	"recipes-desk/pkg/testutils"
)

type TestLogSent struct {
	Message     string
	ServerBlown bool
}

var _ rabbitmq.Event = (*TestLogSent)(nil)

func (e TestLogSent) Type() string {
	return "LogSent"
}

func (e TestLogSent) RoutingKey() string {
	return "logs.sent"
}

func setupConnManager(
	t *testing.T,
	container *rabbitmqtc.RabbitMQContainer,
	log *zerolog.Logger,
) *connection.Manager {
	t.Helper()

	ctx := context.Background()
	url, err := container.AmqpURL(ctx)
	require.NoError(t, err)

	connManager := connection.NewManager(url, log)
	require.NoError(t, connManager.Start(ctx))

	return connManager
}

func TestIntegration_Publisher_NewPublisher(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	ctx := context.Background()

	log := zerolog.New(zerolog.NewConsoleWriter())
	connManager := setupConnManager(t, container, &log)
	defer connManager.Stop()

	tests := []struct {
		name         string
		optionsFuncs []publisher.OptionFunc
		setup        func(pool.Pool[*amqp.Channel])
		wantErr      error
	}{
		{
			name: "success",
			optionsFuncs: []publisher.OptionFunc{
				publisher.WithExchangeDeclare,
				publisher.WithExchangeName("test.exchange"),
			},
			setup:   nil,
			wantErr: nil,
		},
		{
			name: "fails - pool get returns error",
			optionsFuncs: []publisher.OptionFunc{
				publisher.WithExchangeDeclare,
				publisher.WithExchangeName("test.exchange"),
			},
			setup: func(pool pool.Pool[*amqp.Channel]) {
				pool.Close(ctx)
			},
			wantErr: rabbitmq.ErrChannelPoolClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chanpool, err := rabbitmq.NewChannelPool(ctx, connManager, 5, &log)
			require.NoError(t, err)
			defer chanpool.Close(ctx)

			if tt.setup != nil {
				tt.setup(chanpool)
			}

			publisher, err := publisher.New(ctx, chanpool, &log, tt.optionsFuncs...)
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

	ctx := context.Background()

	log := zerolog.New(zerolog.NewConsoleWriter())
	connManager := setupConnManager(t, container, &log)
	defer connManager.Stop()

	tests := []struct {
		name        string
		optionFuncs []publisher.OptionFunc
		setup       func(pool.Pool[*amqp.Channel])
		wantErr     bool
	}{
		{
			name: "successfully publishes event",
			optionFuncs: []publisher.OptionFunc{
				publisher.WithExchangeDeclare,
				publisher.WithExchangeName("test.exchange"),
			},
			setup:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolchan, err := rabbitmq.NewChannelPool(ctx, connManager, 5, &log)
			require.NoError(t, err)
			defer poolchan.Close(ctx)

			publisher, err := publisher.New(ctx, poolchan, &log, tt.optionFuncs...)
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup(poolchan)
			}

			err = publisher.Publish(ctx, TestLogSent{
				Message:     "Test message",
				ServerBlown: true,
			})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPublisher_PublishWithConfirm(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	ctx := context.Background()

	log := zerolog.New(zerolog.NewConsoleWriter())
	connManager := setupConnManager(t, container, &log)
	defer connManager.Stop()

	tests := []struct {
		name        string
		optionFuncs []publisher.OptionFunc
		wantErr     bool
	}{
		{
			name: "successfully publishes with confirmation",
			optionFuncs: []publisher.OptionFunc{
				publisher.WithExchangeDeclare,
				publisher.WithExchangeName("test.exchange"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolchan, err := rabbitmq.NewChannelPool(ctx, connManager, 5, &log)
			require.NoError(t, err)
			defer poolchan.Close(ctx)

			publisher, err := publisher.New(ctx, poolchan, &log, tt.optionFuncs...)
			require.NoError(t, err)

			err = publisher.PublishWithConfirm(ctx, TestLogSent{
				Message:     "Test message",
				ServerBlown: true,
			})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
