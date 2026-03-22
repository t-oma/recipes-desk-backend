//go:build integration

package rabbitmq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/infra/messagebus/rabbitmq"
	"recipes-desk/pkg/testutils"
)

func setupTopology(
	t *testing.T,
	host string,
	port int,
) (*rabbitmq.Connection, *rabbitmq.Topology, func()) {
	t.Helper()

	logger := testutils.NewTestLogger(t, 1)
	config := testConnectionConfig(host, port)

	conn, err := rabbitmq.NewConnection(config, logger)
	require.NoError(t, err)

	topology := rabbitmq.NewTopology(conn, logger)

	cleanup := func() {
		if err = conn.Close(); err != nil {
			t.Logf("Failed to close RabbitMQ connection: %v", err)
		}
	}

	return conn, topology, cleanup
}

func TestIntegration_Topology_DeclareExchange(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	tests := []struct {
		name         string
		exchangeName string
		exchangeType rabbitmq.ExchangeType
	}{
		{
			name:         "success - topic exchange",
			exchangeName: "test.topic",
			exchangeType: rabbitmq.ExchangeTypeTopic,
		},
		{
			name:         "success - direct exchange",
			exchangeName: "test.direct",
			exchangeType: rabbitmq.ExchangeTypeDirect,
		},
		{
			name:         "success - fanout exchange",
			exchangeName: "test.fanout",
			exchangeType: rabbitmq.ExchangeTypeFanout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
			defer cleanupTopology()

			cfg := rabbitmq.NewExchangeConfig(tt.exchangeName, tt.exchangeType)
			err := topology.DeclareExchange(cfg)
			require.NoError(t, err)

			// Verify exchange exists
			exists, err := topology.ExchangeExists(tt.exchangeName)
			require.NoError(t, err)
			assert.True(t, exists, "exchange should exist")
		})
	}
}

func TestIntegration_Topology_DeclareQueue(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	tests := []struct {
		name      string
		queueName string
		queueType rabbitmq.QueueType
	}{
		{
			name:      "success - classic queue",
			queueName: "test.classic-queue",
			queueType: rabbitmq.QueueTypeClassic,
		},
		{
			name:      "success - quorum queue",
			queueName: "test.quorum-queue",
			queueType: rabbitmq.QueueTypeQuorum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
			defer cleanupTopology()

			cfg := rabbitmq.NewQueueConfig(tt.queueName, tt.queueType)
			q, err := topology.DeclareQueue(cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.queueName, q.Name)

			// Verify queue exists
			exists, err := topology.QueueExists(tt.queueName)
			require.NoError(t, err)
			assert.True(t, exists, "queue should exist")
		})
	}
}

func TestIntegration_Topology_BindQueue(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		exchangeCfg := rabbitmq.NewExchangeConfig("test.bind.exchange", rabbitmq.ExchangeTypeTopic)
		err := topology.DeclareExchange(exchangeCfg)
		require.NoError(t, err)

		queueCfg := rabbitmq.NewQueueConfig("test.bind-queue", rabbitmq.QueueTypeClassic)
		_, err = topology.DeclareQueue(queueCfg)
		require.NoError(t, err)

		bindingCfg := rabbitmq.NewBindingConfig(
			queueCfg.Name,
			exchangeCfg.Name,
			"test.routing.key",
		)
		err = topology.BindQueue(bindingCfg)
		require.NoError(t, err)

		// Verify exchange and queue still exist
		exists, err := topology.ExchangeExists(exchangeCfg.Name)
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = topology.QueueExists(queueCfg.Name)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestIntegration_Topology_SetupTopology(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		exchangeCfg := rabbitmq.NewExchangeConfig(
			"test.topology.exchange",
			rabbitmq.ExchangeTypeTopic,
		)
		queueCfg := rabbitmq.NewQueueConfig("test.topology-queue", rabbitmq.QueueTypeClassic)
		bindingCfg := rabbitmq.NewBindingConfig(
			queueCfg.Name,
			exchangeCfg.Name,
			"test.topology.key",
		)

		err := topology.SetupTopology(exchangeCfg, queueCfg, bindingCfg)
		require.NoError(t, err)

		// Verify exchange and queue exist
		exists, err := topology.ExchangeExists(exchangeCfg.Name)
		require.NoError(t, err)
		assert.True(t, exists, "exchange should exist")

		exists, err = topology.QueueExists(queueCfg.Name)
		require.NoError(t, err)
		assert.True(t, exists, "queue should exist")
	})
}

func TestIntegration_Topology_SetupDLQ(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		queueCfg := rabbitmq.NewQueueConfig("test.dlq-queue", rabbitmq.QueueTypeClassic)
		result, err := topology.SetupDLQ(queueCfg)
		require.NoError(t, err)

		assert.NotNil(t, result.Args)
		assert.Equal(t, queueCfg.Name+".dlx", result.Args["x-dead-letter-exchange"])

		// Verify DLX and DLQ exist
		exists, err := topology.ExchangeExists(queueCfg.Name + ".dlx")
		require.NoError(t, err)
		assert.True(t, exists, "DLX should exist")

		exists, err = topology.QueueExists(queueCfg.Name + ".dlq")
		require.NoError(t, err)
		assert.True(t, exists, "DLQ should exist")
	})
}

func TestIntegration_Topology_SetupTopologyWithDLQ(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		exchangeCfg := rabbitmq.NewExchangeConfig("test.full.exchange", rabbitmq.ExchangeTypeTopic)
		queueCfg := rabbitmq.NewQueueConfig("test.full-queue", rabbitmq.QueueTypeClassic)
		bindingCfg := rabbitmq.NewBindingConfig(
			queueCfg.Name,
			exchangeCfg.Name,
			"test.full.key",
		)

		err := topology.SetupTopologyWithDLQ(exchangeCfg, queueCfg, bindingCfg)
		require.NoError(t, err)

		// Verify main exchange and queue exist
		exists, err := topology.ExchangeExists(exchangeCfg.Name)
		require.NoError(t, err)
		assert.True(t, exists, "main exchange should exist")

		exists, err = topology.QueueExists(queueCfg.Name)
		require.NoError(t, err)
		assert.True(t, exists, "main queue should exist")

		// Verify DLX and DLQ exist
		exists, err = topology.ExchangeExists(queueCfg.Name + ".dlx")
		require.NoError(t, err)
		assert.True(t, exists, "DLX should exist")

		exists, err = topology.QueueExists(queueCfg.Name + ".dlq")
		require.NoError(t, err)
		assert.True(t, exists, "DLQ should exist")
	})
}

func TestIntegration_Topology_DeleteExchange(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		exchangeCfg := rabbitmq.NewExchangeConfig(
			"test.delete.exchange",
			rabbitmq.ExchangeTypeTopic,
		)
		err := topology.DeclareExchange(exchangeCfg)
		require.NoError(t, err)

		err = topology.DeleteExchange(exchangeCfg.Name, false)
		require.NoError(t, err)

		// Verify exchange no longer exists
		exists, err := topology.ExchangeExists(exchangeCfg.Name)
		require.NoError(t, err)
		assert.False(t, exists, "exchange should not exist after deletion")
	})
}

func TestIntegration_Topology_DeleteQueue(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		queueCfg := rabbitmq.NewQueueConfig("test.delete-queue", rabbitmq.QueueTypeClassic)
		_, err := topology.DeclareQueue(queueCfg)
		require.NoError(t, err)

		err = topology.DeleteQueue(queueCfg.Name, false, false)
		require.NoError(t, err)

		// Verify queue no longer exists
		exists, err := topology.QueueExists(queueCfg.Name)
		require.NoError(t, err)
		assert.False(t, exists, "queue should not exist after deletion")
	})
}

func TestIntegration_Topology_PurgeQueue(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success - empty queue", func(t *testing.T) {
		_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
		defer cleanupTopology()

		queueCfg := rabbitmq.NewQueueConfig("test.purge-queue", rabbitmq.QueueTypeClassic)
		_, err := topology.DeclareQueue(queueCfg)
		require.NoError(t, err)

		count, err := topology.PurgeQueue(queueCfg.Name)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Verify queue still exists after purge
		exists, err := topology.QueueExists(queueCfg.Name)
		require.NoError(t, err)
		assert.True(t, exists, "queue should still exist after purge")
	})
}

func TestIntegration_Topology_ExchangeExists(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	tests := []struct {
		name          string
		setupExchange func(t *testing.T, topology *rabbitmq.Topology)
		exchangeName  string
		wantExists    bool
	}{
		{
			name: "exists - after declaration",
			setupExchange: func(t *testing.T, topology *rabbitmq.Topology) {
				cfg := rabbitmq.NewExchangeConfig("test.exists.exchange", rabbitmq.ExchangeTypeTopic)
				err := topology.DeclareExchange(cfg)
				require.NoError(t, err)
			},
			exchangeName: "test.exists.exchange",
			wantExists:   true,
		},
		{
			name:          "not exists - never declared",
			setupExchange: func(_ *testing.T, _ *rabbitmq.Topology) {},
			exchangeName:  "test.nonexistent.exchange",
			wantExists:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
			defer cleanupTopology()

			tt.setupExchange(t, topology)

			exists, err := topology.ExchangeExists(tt.exchangeName)
			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

func TestIntegration_Topology_QueueExists(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	tests := []struct {
		name       string
		setupQueue func(t *testing.T, topology *rabbitmq.Topology)
		queueName  string
		wantExists bool
	}{
		{
			name: "exists - after declaration",
			setupQueue: func(t *testing.T, topology *rabbitmq.Topology) {
				cfg := rabbitmq.NewQueueConfig("test.exists.queue", rabbitmq.QueueTypeClassic)
				_, err := topology.DeclareQueue(cfg)
				require.NoError(t, err)
			},
			queueName:  "test.exists.queue",
			wantExists: true,
		},
		{
			name:       "not exists - never declared",
			setupQueue: func(_ *testing.T, _ *rabbitmq.Topology) {},
			queueName:  "test.nonexistent.queue",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
			defer cleanupTopology()

			tt.setupQueue(t, topology)

			exists, err := topology.QueueExists(tt.queueName)
			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}
