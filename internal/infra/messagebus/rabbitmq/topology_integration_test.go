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
		exchaneName  string
		exchangeType rabbitmq.ExchangeType
	}{
		{
			name:         "success - topic exchange",
			exchaneName:  "test.topic",
			exchangeType: rabbitmq.ExchangeTypeTopic,
		},
		{
			name:         "success - direct exchange",
			exchaneName:  "test.direct",
			exchangeType: rabbitmq.ExchangeTypeDirect,
		},
		{
			name:         "success - fanout exchange",
			exchaneName:  "test.fanout",
			exchangeType: rabbitmq.ExchangeTypeFanout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, topology, cleanupTopology := setupTopology(t, container.Host, container.Port)
			defer cleanupTopology()

			cfg := rabbitmq.NewExchangeConfig(tt.exchaneName, tt.exchangeType)
			err := topology.DeclareExchange(cfg)
			require.NoError(t, err)
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
	})
}
