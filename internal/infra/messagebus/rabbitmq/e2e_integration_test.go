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

	producerConfig := rabbitmq.DefaultProducerConfig()
	producer, err := rabbitmq.NewProducer(conn, testutils.NewTestLogger(t), producerConfig)
	require.NoError(t, err)
	defer producer.Close()

	consumer, err := rabbitmq.NewConsumer(conn, log)
	require.NoError(t, err)
	defer consumer.Close()

	tests := []struct {
		name           string
		queueType      rabbitmq.QueueType
		queueName      string
		routingKey     string
		publishings    int
		handler        func(msg messagebus.Message) error
		handlerTimeout time.Duration
	}{
		{
			name:        "classic queue",
			queueType:   rabbitmq.QueueTypeClassic,
			queueName:   "test.e2e.classic",
			routingKey:  "test.key",
			publishings: 3,
			handler:     func(_ messagebus.Message) error { return nil },
		},
		{
			name:        "quorum queue",
			queueType:   rabbitmq.QueueTypeQuorum,
			queueName:   "test.e2e.quorum",
			routingKey:  "test.key",
			publishings: 3,
			handler:     func(_ messagebus.Message) error { return nil },
		},
		{
			name:        "handler error",
			queueType:   rabbitmq.QueueTypeClassic,
			queueName:   "test.e2e.error",
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
			name:        "handler timeout",
			queueType:   rabbitmq.QueueTypeClassic,
			queueName:   "test.e2e.timeout",
			routingKey:  "test.key",
			publishings: 5,
			handler: func(_ messagebus.Message) error {
				if rand.IntN(2) == 0 {
					time.Sleep(300 * time.Millisecond)
				}

				return nil
			},
			handlerTimeout: 100 * time.Millisecond,
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
			handler := func(msg messagebus.Message) error {
				wg.Add(1)
				defer func() {
					wg.Done()
				}()

				if err := tt.handler(msg); err != nil {
					return err
				}

				return nil
			}

			// Start consumer
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			go func() {
				if err := consumer.Consume(ctx, queueCfg.Name, handler); err != nil {
					t.Logf("Consumer failed: %v", err)
				}
			}()

			// Wait for consumer to start
			time.Sleep(100 * time.Millisecond)

			for i := 0; i < tt.publishings; i++ {
				msg := messagebus.Message{
					ID:      "id" + fmt.Sprintf("%06d", i),
					Type:    "TestMessage",
					Payload: json.RawMessage(`{"test": "data"}`),
				}
				err = producer.Publish(ctx, exchangeCfg.Name, tt.routingKey, msg)
				require.NoError(t, err)
			}

			wg.Wait()
		})
	}
}
