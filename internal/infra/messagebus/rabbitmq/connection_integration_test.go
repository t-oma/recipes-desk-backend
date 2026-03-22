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

func setupConnection(
	t *testing.T,
	host string,
	port int,
) (*rabbitmq.Connection, func()) {
	logger := testutils.NewTestLogger(t)
	config := testConnectionConfig(host, port)
	conn, err := rabbitmq.NewConnection(config, logger)
	require.NoError(t, err)
	return conn, func() {
		if err = conn.Close(); err != nil {
			t.Logf("Failed to close RabbitMQ connection: %v", err)
		}
	}
}

func TestIntegration_Connection_NewConnection(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		conn, cleanupConn := setupConnection(t, container.Host, container.Port)
		defer cleanupConn()

		assert.True(t, conn.IsConnected())
	})
}

func TestIntegration_Connection_Channel(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		conn, cleanupConn := setupConnection(t, container.Host, container.Port)
		defer cleanupConn()

		ch, err := conn.Channel()
		require.NoError(t, err)
		defer ch.Close()

		assert.NotNil(t, ch)
	})
}

func TestIntegration_Connection_Channel_NotConnected(t *testing.T) {
	// Don't create connection - test with nil connection scenario
	cfg := testConnectionConfig("unexisting-host", 9999)
	cfg.ConnectionTimeout = 1 * time.Second

	// This should fail because host doesn't exist
	logger := testutils.NewTestLogger(t)
	_, err := rabbitmq.NewConnection(cfg, logger)
	require.Error(t, err)
}

func TestIntegration_Connection_Close(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	t.Run("success", func(t *testing.T) {
		conn, _ := setupConnection(t, container.Host, container.Port)
		require.True(t, conn.IsConnected())

		err := conn.Close()
		require.NoError(t, err)

		assert.False(t, conn.IsConnected())
	})

	t.Run("success - close multiple times", func(t *testing.T) {
		conn, cleanupConn := setupConnection(t, container.Host, container.Port)
		// Call cleanup (should be safe to call after Close)
		defer cleanupConn()
		require.True(t, conn.IsConnected())

		err := conn.Close()
		require.NoError(t, err)

		assert.False(t, conn.IsConnected())
	})
}

func TestIntegration_Connection_Ping(t *testing.T) {
	container, cleanup := testutils.SetupRabbitMQContainer(t)
	defer cleanup()

	tests := []struct {
		name  string
		setup func(t *testing.T, conn *rabbitmq.Connection)
		want  error
	}{
		{
			name: "success",
			setup: func(_ *testing.T, _ *rabbitmq.Connection) {
			},
			want: nil,
		},
		{
			name: "connection closed",
			setup: func(t *testing.T, conn *rabbitmq.Connection) {
				err := conn.Close()
				require.NoError(t, err)
			},
			want: rabbitmq.ErrConnectionClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, cleanup := setupConnection(t, container.Host, container.Port)
			defer cleanup()

			tt.setup(t, conn)

			err := conn.Ping(context.Background())
			if tt.want != nil {
				require.Error(t, err)
				assert.ErrorIs(t, tt.want, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
