//go:build integration

package rabbitmq_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/infra/messagebus"
	"recipes-desk/internal/infra/messagebus/rabbitmq"
	"recipes-desk/pkg/testutils"
)

func TestIntegration_E2E_RoundTrip(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	log := testutils.NewTestLogger(t, 1)
	conn, cleanupConn := setupConnection(t, log, container.Host, container.Port)
	defer cleanupConn()

	// Disable logging for topology
	topology := rabbitmq.NewTopology(conn, testutils.NewTestLogger(t))

	producer, err := rabbitmq.NewProducer(
		conn,
		testutils.NewTestLogger(t),
		rabbitmq.DefaultProducerConfig(),
	)
	require.NoError(t, err)
	defer producer.Close()

	tests := []struct {
		name        string
		queueType   rabbitmq.QueueType
		queueName   string
		routingKey  string
		publishings int
		handler     func(msg messagebus.Message) error
	}{
		{
			name:        "classic queue",
			queueType:   rabbitmq.QueueTypeClassic,
			queueName:   "test.e2e-classic",
			routingKey:  "test.key",
			publishings: 3,
			handler:     func(_ messagebus.Message) error { return nil },
		},
		{
			name:        "quorum queue",
			queueType:   rabbitmq.QueueTypeQuorum,
			queueName:   "test.e2e-quorum",
			routingKey:  "test.key",
			publishings: 3,
			handler:     func(_ messagebus.Message) error { return nil },
		},
		{
			name:        "handler error",
			queueType:   rabbitmq.QueueTypeClassic,
			queueName:   "test.e2e-error",
			routingKey:  "test.key",
			publishings: 5,
			handler: func(_ messagebus.Message) error {
				if rand.IntN(2) == 0 {
					return errors.New("handler error")
				}

				return nil
			},
		},
		{
			name:        "handler panic",
			queueType:   rabbitmq.QueueTypeClassic,
			queueName:   "test.e2e-panic",
			routingKey:  "test.key",
			publishings: 5,
			handler: func(_ messagebus.Message) error {
				if rand.IntN(2) == 0 {
					panic("handler panic")
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchangeCfg := rabbitmq.NewExchangeConfig(
				"test.e2e.exchange",
				rabbitmq.ExchangeTypeTopic,
			)
			queueCfg := rabbitmq.NewQueueConfig(tt.queueName, tt.queueType)
			bindingCfg := rabbitmq.NewBindingConfig(queueCfg.Name, exchangeCfg.Name, tt.routingKey)

			err := topology.SetupTopology(exchangeCfg, queueCfg, bindingCfg)
			require.NoError(t, err)

			var wg sync.WaitGroup
			wg.Add(tt.publishings)
			handler := func(msg messagebus.Message) error {
				defer wg.Done()

				return tt.handler(msg)
			}

			consumer, err := rabbitmq.NewConsumer(conn, log)
			require.NoError(t, err)
			defer consumer.Close()
			go func() {
				if err := consumer.Consume(queueCfg.Name, handler); err != nil {
					t.Logf("Consumer failed: %v", err)
				}
			}()

			// Wait for consumer to start
			time.Sleep(100 * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			for i := 0; i < tt.publishings; i++ {
				msg := messagebus.Message{
					ID:      "id" + fmt.Sprintf("%06d", i),
					Type:    "TestMessage",
					Payload: json.RawMessage(`{"test": "data"}`),
				}
				err = producer.Publish(ctx, exchangeCfg.Name, tt.routingKey, msg)
				require.NoError(t, err)
			}

			wg.Wait() // Wait for all messages to be processed
			cancel()
		})
	}
}

func TestIntegration_E2E_DeadLetterQueue(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()
	log := testutils.NewTestLogger(t, 1)
	conn, cleanupConn := setupConnection(t, log, container.Host, container.Port)
	defer cleanupConn()

	topology := rabbitmq.NewTopology(conn, testutils.NewTestLogger(t))

	producer, err := rabbitmq.NewProducer(
		conn,
		testutils.NewTestLogger(t),
		rabbitmq.DefaultProducerConfig(),
	)
	require.NoError(t, err)
	defer producer.Close()

	t.Run("success - empty queue", func(t *testing.T) {
		exchangeCfg := rabbitmq.NewExchangeConfig("test.dlq.exchange", rabbitmq.ExchangeTypeTopic)
		queueCfg := rabbitmq.NewQueueConfig("test.dlq-queue", rabbitmq.QueueTypeClassic)
		bindingCfg := rabbitmq.NewBindingConfig(queueCfg.Name, exchangeCfg.Name, "test.key")
		err := topology.SetupTopologyWithDLQ(exchangeCfg, queueCfg, bindingCfg)
		require.NoError(t, err)

		dlqExists, err := topology.QueueExists(queueCfg.Name + ".dlq")
		require.NoError(t, err)
		assert.True(t, dlqExists, "DLQ should exist")

		failedCount := 0
		var mu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(1)
		handler := func(msg messagebus.Message) error {
			defer wg.Done()
			mu.Lock()
			failedCount++
			mu.Unlock()
			return errors.New("always fail")
		}

		consumer, err := rabbitmq.NewConsumer(conn, log)
		require.NoError(t, err)
		defer consumer.Close()
		go func() {
			_ = consumer.Consume(queueCfg.Name, handler)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		msg := messagebus.Message{
			Type:    "TestMessage",
			Payload: json.RawMessage(`{"test": "data"}`),
		}
		err = producer.Publish(ctx, exchangeCfg.Name, "test.key", msg)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		wg.Wait() // Wait for all messages to be processed
		cancel()

		ch, err := conn.Channel()
		require.NoError(t, err)
		defer ch.Close()

		_, ok, err := ch.Get(queueCfg.Name, false)
		require.NoError(t, err)
		assert.False(t, ok, "main queue should be empty")

		_, ok, err = ch.Get(queueCfg.Name+".dlq", false)
		require.NoError(t, err)
		assert.True(t, ok, "DLQ should contain the failed message")
	})
}
