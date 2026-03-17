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

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/rabbitmq"
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
	_testTimeout  = 3 * time.Second
)

func setupPublisher(t *testing.T, conn *amqp.Connection) *rabbitmq.Publisher {
	t.Helper()

	log := zerolog.New(zerolog.NewConsoleWriter())
	publisher := rabbitmq.NewPublisher(conn, _testExchange, &log)
	require.NoError(t, publisher.ExchangeDeclare())

	return publisher
}

func setupConsumer(
	t *testing.T,
	conn *amqp.Connection,
	queueName string,
	handler rabbitmq.HandlerFunc,
) *rabbitmq.Consumer {
	t.Helper()

	log := zerolog.New(zerolog.NewConsoleWriter())
	consumer := rabbitmq.NewConsumer(conn, rabbitmq.ConsumerConfig{
		Exchange:   _testExchange,
		Queue:      queueName,
		RoutingKey: "logs.sent",
		Handler:    handler,
		MaxRetries: 3,
		RetryTTLs: []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			500 * time.Millisecond,
		},
	}, &log)

	return consumer
}

func TestIntegration_Publisher_PublishWithConfirm(t *testing.T) {
	conn, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	publisher := setupPublisher(t, conn)

	event := TestLogSent{
		Message:     "Test message",
		ServerBlown: false,
	}

	t.Run("publish with confirmation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), _testTimeout)
		defer cancel()

		err := publisher.PublishWithConfirm(ctx, event)
		require.NoError(t, err)
	})

	t.Run("publish without confirmation", func(t *testing.T) {
		err := publisher.Publish(context.Background(), event)
		require.NoError(t, err)
	})
}

func TestIntegration_Consumer_BasicConsume(t *testing.T) {
	conn, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	publisher := setupPublisher(t, conn)

	received := make(chan TestLogSent, 1)
	handler := func(ctx context.Context, routingKey string, body []byte) error {
		var event TestLogSent
		if err := json.Unmarshal(body, &event); err != nil {
			return err
		}
		received <- event
		return nil
	}

	consumer := setupConsumer(t, conn, _testQueue, handler)
	require.NoError(t, consumer.Setup())

	ctx, cancel := context.WithTimeout(context.Background(), _testTimeout)
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			t.Logf("Consumer error: %v", err)
		}
	}()

	// Give consumer time to start
	time.Sleep(10 * time.Millisecond)

	event := TestLogSent{
		Message:     "Test message",
		ServerBlown: true,
	}

	err := publisher.PublishWithConfirm(ctx, event)
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
	conn, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	publisher := setupPublisher(t, conn)

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

	consumer := setupConsumer(t, conn, _testQueue, handler)
	require.NoError(t, consumer.Setup())

	ctx, cancel := context.WithTimeout(context.Background(), _testTimeout)
	defer cancel()

	go func() {
		if err := consumer.Start(ctx); err != nil {
			t.Logf("Consumer error: %v", err)
		}
	}()

	// Give consumer time to start
	time.Sleep(10 * time.Millisecond)

	event := TestLogSent{
		Message:     "Test message",
		ServerBlown: false,
	}

	err := publisher.PublishWithConfirm(ctx, event)
	require.NoError(t, err)

	// Wait for all retry attempts
	wg.Wait()
	assert.Equal(t, 3, attempts)
}
