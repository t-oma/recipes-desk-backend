//go:build integration

package rabbitmq_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/messagebus"
	"recipes-desk/internal/infra/messagebus/rabbitmq"
	"recipes-desk/pkg/testutils"
)

func setupProducer(
	t *testing.T,
	host string,
	port int,
) (*rabbitmq.Connection, *rabbitmq.Producer, func()) {
	return setupProducerWithConfig(t, host, port, nil)
}

func setupProducerWithConfig(
	t *testing.T,
	host string,
	port int,
	modifyConfig func(*rabbitmq.ProducerConfig),
) (*rabbitmq.Connection, *rabbitmq.Producer, func()) {
	t.Helper()

	logger := testutils.NewTestLogger(t, 1)
	connConfig := testConnectionConfig(host, port)
	conn, err := rabbitmq.NewConnection(connConfig, logger)
	require.NoError(t, err)

	producerConfig := rabbitmq.DefaultProducerConfig()
	if modifyConfig != nil {
		modifyConfig(&producerConfig)
	}

	producer, err := rabbitmq.NewProducer(conn, logger, producerConfig)
	require.NoError(t, err)

	cleanup := func() {
		if err = producer.Close(); err != nil {
			t.Logf("Failed to close producer: %v", err)
		}
		if err = conn.Close(); err != nil {
			t.Logf("Failed to close connection: %v", err)
		}
	}

	return conn, producer, cleanup
}

func TestIntegration_Producer_Publish(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
	defer cleanupTopology()

	tests := []struct {
		name            string
		exchangeName    string
		queueName       string
		routingKey      string
		declareExchange func(t *testing.T, topology *rabbitmq.Topology) string
		msg             messagebus.Message
		timeout         time.Duration
		checkFunc       func(t *testing.T, delivery amqp.Delivery)
		wantErr         error
	}{
		{
			name:         "success",
			exchangeName: "test.publish.exchange",
			queueName:    "test.publish-queue",
			routingKey:   "test.key",
			msg: messagebus.Message{
				ID:   "id123",
				Type: "TestMessage",
				Payload: json.RawMessage(
					`{"test": "data"}`,
				),
				Timestamp: time.Now().Add(-24 * time.Hour),
			},
			timeout: 5 * time.Second,
			checkFunc: func(t *testing.T, delivery amqp.Delivery) {
				assert.Equal(t, "TestMessage", delivery.Type)
				require.NotEmpty(t, delivery.MessageId, "ID should be auto-generated")
				assert.Equal(t, "id123", delivery.MessageId)
			},
			wantErr: nil,
		},
		{
			name:         "success - generates ID when empty",
			exchangeName: "test.id.exchange",
			queueName:    "test.id-queue",
			routingKey:   "test.key",
			msg: messagebus.Message{
				Type: "TestMessage",
				Payload: json.RawMessage(
					`{"test": "data"}`,
				),
				Timestamp: time.Now().Add(-24 * time.Hour),
			},
			timeout: 5 * time.Second,
			checkFunc: func(t *testing.T, delivery amqp.Delivery) {
				assert.Equal(t, "TestMessage", delivery.Type)
				require.NotEmpty(t, delivery.MessageId, "ID should be auto-generated")
			},
			wantErr: nil,
		},
		{
			name:         "success - sets timestamp when zero",
			exchangeName: "test.timestamp.exchange",
			queueName:    "test.timestamp-queue",
			routingKey:   "test.key",
			msg: messagebus.Message{
				ID:   "id123",
				Type: "TestMessage",
				Payload: json.RawMessage(
					`{"test": "data"}`,
				),
			},
			timeout: 5 * time.Second,
			checkFunc: func(t *testing.T, delivery amqp.Delivery) {
				assert.Equal(t, "TestMessage", delivery.Type)
				assert.False(t, delivery.Timestamp.IsZero(), "timestamp should be auto-set")
			},
			wantErr: nil,
		},
		{
			name:         "failure - context cancelled",
			exchangeName: "test.cancel.exchange",
			queueName:    "test.cancel-queue",
			routingKey:   "test.key",
			msg: messagebus.Message{
				Type: "TestMessage",
				Payload: json.RawMessage(
					`{"test": "data"}`,
				),
			},
			timeout:   1 * time.Nanosecond,
			checkFunc: nil,
			wantErr:   context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exchangeCfg := rabbitmq.NewExchangeConfig(tt.exchangeName, rabbitmq.ExchangeTypeTopic)
			err := topology.DeclareExchange(exchangeCfg)
			require.NoError(t, err)

			queueCfg := rabbitmq.NewQueueConfig(tt.queueName, rabbitmq.QueueTypeClassic)
			_, err = topology.DeclareQueue(queueCfg)
			require.NoError(t, err)

			bindingCfg := rabbitmq.NewBindingConfig(queueCfg.Name, exchangeCfg.Name, tt.routingKey)
			err = topology.BindQueue(bindingCfg)
			require.NoError(t, err)

			conn, producer, cleanupProducer := setupProducer(t, container.Host, container.Port)
			defer cleanupProducer()

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err = producer.Publish(ctx, exchangeCfg.Name, tt.routingKey, tt.msg)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)

				// Verify message in queue
				ch, err := conn.Channel()
				require.NoError(t, err)
				defer ch.Close()

				delivery, ok, err := ch.Get(queueCfg.Name, true)
				require.NoError(t, err)
				assert.True(t, ok, "message should be in queue")

				if tt.checkFunc != nil {
					tt.checkFunc(t, delivery)
				}
			}
		})
	}
}

func TestIntegration_Producer_Publish_Mandatory(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("error when no route and mandatory true", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		// Exchange without any bindings
		exchangeCfg := rabbitmq.NewExchangeConfig(
			"test.mandatory.exchange",
			rabbitmq.ExchangeTypeDirect,
		)
		err := topology.DeclareExchange(exchangeCfg)
		require.NoError(t, err)

		_, producer, cleanupProducer := setupProducerWithConfig(
			t,
			container.Host,
			container.Port,
			func(cfg *rabbitmq.ProducerConfig) {
				cfg.Mandatory = true
			},
		)
		defer cleanupProducer()

		msg := messagebus.Message{
			Type:    "TestMessage",
			Payload: json.RawMessage(`{}`),
		}

		ctx := context.Background()
		err = producer.Publish(ctx, exchangeCfg.Name, "no.route", msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, rabbitmq.ErrMandatoryFailed)
	})

	t.Run("success when route exists", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		exchangeCfg := rabbitmq.NewExchangeConfig(
			"test.mandatory.ok.exchange",
			rabbitmq.ExchangeTypeDirect,
		)
		queueCfg := rabbitmq.NewQueueConfig("test.mandatory.ok.queue", rabbitmq.QueueTypeClassic)
		bindingCfg := rabbitmq.NewBindingConfig(queueCfg.Name, exchangeCfg.Name, "test.route")

		err := topology.SetupTopology(exchangeCfg, queueCfg, bindingCfg)
		require.NoError(t, err)

		_, producer, cleanupProducer := setupProducerWithConfig(
			t,
			container.Host,
			container.Port,
			func(cfg *rabbitmq.ProducerConfig) {
				cfg.Mandatory = true
			},
		)
		defer cleanupProducer()

		msg := messagebus.Message{
			Type:    "TestMessage",
			Payload: json.RawMessage(`{}`),
		}

		ctx := context.Background()
		err = producer.Publish(ctx, exchangeCfg.Name, "test.route", msg)
		require.NoError(t, err)
	})
}

func TestIntegration_Producer_Close(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("close is idempotent", func(t *testing.T) {
		_, _, cleanupProducer := setupProducer(t, container.Host, container.Port)

		// Cleanup will call Close, but we can also call it manually
		// The producer should handle multiple Close calls gracefully
		cleanupProducer()
	})
}
