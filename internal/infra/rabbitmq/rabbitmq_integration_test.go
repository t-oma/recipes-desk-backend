//go:build integration
// +build integration

package rabbitmq_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rabbitmqtc "github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	"recipes-desk/internal/infra/rabbitmq"
	"recipes-desk/internal/infra/rabbitmq/connection"
	"recipes-desk/internal/infra/rabbitmq/consumer"
	"recipes-desk/internal/infra/rabbitmq/publisher"
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

const (
	_testExchange = "test.logs.events"
	_testQueue    = "test.logs-sent"
	_testTimeout  = 5 * time.Second
)

// setupRabbitMQFromContainer creates ConnectionManager from container.
func setupRabbitMQFromContainer(
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

func setupPublisher(
	t *testing.T,
	pool *rabbitmq.ChannelPool,
	log *zerolog.Logger,
) *publisher.Publisher {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), _testTimeout)
	defer cancel()

	publisher, err := publisher.New(
		ctx,
		pool,
		log,
		publisher.WithExchangeDeclare,
		publisher.WithExchangeName(_testExchange),
	)

	require.NoError(t, err)

	return publisher
}

func setupConsumer(
	t *testing.T,
	pool *rabbitmq.ChannelPool,
	handler consumer.HandlerFunc,
	log *zerolog.Logger,
) *consumer.Consumer {
	t.Helper()

	// consumer.Config{
	// 	Exchange:   _testExchange,
	// 	Queue:      _testQueue,
	// 	RoutingKey: "logs.sent",
	// 	Handler:    handler,
	// 	MaxRetries: 3,
	// 	RetryTTLs: []time.Duration{
	// 		100 * time.Millisecond,
	// 		200 * time.Millisecond,
	// 		500 * time.Millisecond,
	// 	},
	// }

	consumer, err := consumer.New(context.Background(),
		_testQueue,
		pool,
		log,
		consumer.WithExchangeName(_testExchange),
		consumer.WithExchangeKind("topic"),
		consumer.WithRoutingKey("logs.sent"),
		consumer.WithHandler(handler),
		consumer.WithMaxRetries(3),
		consumer.WithRetryTTLs([]time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			500 * time.Millisecond,
		}),
	)
	require.NoError(t, err)

	return consumer
}

func TestIntegration_Consumer_BasicConsume(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	log := zerolog.New(zerolog.NewConsoleWriter())
	connManager := setupRabbitMQFromContainer(t, container, &log)
	defer connManager.Stop()

	pool, err := rabbitmq.NewChannelPool(context.Background(), connManager, 5, &log)
	require.NoError(t, err)
	defer pool.Close(context.Background())

	publisher := setupPublisher(t, pool, &log)

	received := make(chan TestLogSent, 1)
	handler := func(ctx context.Context, routingKey string, body []byte) error {
		var event TestLogSent
		if err := json.Unmarshal(body, &event); err != nil {
			return err
		}
		received <- event
		return nil
	}

	consumer := setupConsumer(t, pool, handler, &log)
	require.NoError(t, consumer.Setup(context.Background()))
	defer func() {
		if err := consumer.Shutdown(context.Background()); err != nil {
			t.Logf("Consumer shutdown error: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), _testTimeout)
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			t.Logf("Consumer error: %v", err)
		}
	}()

	// Give consumer time to start
	time.Sleep(10 * time.Millisecond)

	// Publish event
	event := TestLogSent{
		Message:     "Test message",
		ServerBlown: true,
	}

	err = publisher.PublishWithConfirm(ctx, event)
	require.NoError(t, err)

	// Wait for message
	select {
	case receivedEvent := <-received:
		assert.Equal(t, event.Message, receivedEvent.Message)
		assert.Equal(t, event.ServerBlown, receivedEvent.ServerBlown)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestIntegration_Consumer_Retry(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	log := zerolog.New(zerolog.NewConsoleWriter())
	connManager := setupRabbitMQFromContainer(t, container, &log)
	defer connManager.Stop()

	pool, err := rabbitmq.NewChannelPool(context.Background(), connManager, 5, &log)
	require.NoError(t, err)
	defer pool.Close(context.Background())

	publisher := setupPublisher(t, pool, &log)

	wg := sync.WaitGroup{}
	wg.Add(3)
	// Setup consumer that fails first 2 times
	attempts := 0
	handler := func(ctx context.Context, routingKey string, body []byte) error {
		defer wg.Done()
		attempts++
		if attempts < 3 {
			return assert.AnError
		}
		return nil
	}

	consumer := setupConsumer(t, pool, handler, &log)
	require.NoError(t, consumer.Setup(context.Background()))
	defer func() {
		if err := consumer.Shutdown(context.Background()); err != nil {
			t.Logf("Consumer shutdown error: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), _testTimeout)
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			t.Logf("Consumer error: %v", err)
		}
	}()

	// Give consumer time to start
	time.Sleep(10 * time.Millisecond)

	// Publish event
	event := TestLogSent{
		Message:     "Test message",
		ServerBlown: false,
	}

	err = publisher.PublishWithConfirm(ctx, event)
	require.NoError(t, err)

	// Wait for all retry attempts
	wg.Wait()
	assert.Equal(t, 3, attempts)
}
