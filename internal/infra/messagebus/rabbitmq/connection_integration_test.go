//go:build integration
// +build integration

package rabbitmq_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"recipes-desk/internal/infra/messagebus/rabbitmq"
	"recipes-desk/pkg/testutils"
)

func testConnectionConfig(host string, port int) rabbitmq.ConnectionConfig {
	//nolint:exhaustruct // test struct
	return rabbitmq.ConnectionConfig{
		Host:     host,
		Port:     port,
		User:     "guest",
		Password: "guest",
		VHost:    "/",
		Reconnect: rabbitmq.ReconnectConfig{
			Enabled: false,
		},
	}
}

func setupConnection(t *testing.T) (*rabbitmq.Connection, func()) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)

	logger := testutils.NewTestLogger(t)
	config := testConnectionConfig(container.Host, container.Port)
	conn, err := rabbitmq.NewConnection(config, logger)
	require.NoError(t, err)
	return conn, func() {
		if err = conn.Close(); err != nil {
			t.Logf("Failed to close RabbitMQ connection: %v", err)
		}
		cleanup()
	}
}

func TestIntegration_Connection_NewConnection(t *testing.T) {
	conn, cleanup := setupConnection(t)
	defer cleanup()

	assert.True(t, conn.IsConnected())
}

func TestIntegration_Connection_Channel(t *testing.T) {
	conn, cleanup := setupConnection(t)
	defer cleanup()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	assert.NotNil(t, ch)
}

func TestIntegration_Connection_Channel_NotConnected(t *testing.T) {
	_, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	// Don't create connection - test with nil connection scenario
	cfg := testConnectionConfig("localhost", 5672)
	cfg.ConnectionTimeout = 1 * time.Second

	// This should fail because host doesn't exist
	logger := testutils.NewTestLogger(t)
	_, err := rabbitmq.NewConnection(cfg, logger)
	require.Error(t, err)
}

func TestIntegration_Connection_Close(t *testing.T) {
	conn, cleanup := setupConnection(t)

	require.True(t, conn.IsConnected())

	err := conn.Close()
	require.NoError(t, err)

	assert.False(t, conn.IsConnected())

	// Call cleanup (should be safe to call after Close)
	cleanup()
}

func TestIntegration_Connection_Ping(t *testing.T) {
	conn, cleanup := setupConnection(t)
	defer cleanup()

	err := conn.Ping(context.Background())
	assert.NoError(t, err)
}

func TestIntegration_Connection_Ping_AfterClose(t *testing.T) {
	conn, cleanup := setupConnection(t)
	defer cleanup()

	err := conn.Close()
	require.NoError(t, err)

	err = conn.Ping(context.Background())
	assert.Error(t, err)
}
