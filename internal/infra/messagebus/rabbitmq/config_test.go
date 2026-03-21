package rabbitmq_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"recipes-desk/internal/infra/messagebus/rabbitmq"
)

func TestConnectionConfig_BuildURI(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() rabbitmq.ConnectionConfig
		want        string
	}{
		{
			name: "basic connection without TLS",
			setupConfig: func() rabbitmq.ConnectionConfig {
				cfg := rabbitmq.DefaultConnectionConfig()
				cfg.Host = "localhost"
				cfg.Port = 5672
				cfg.User = "guest"
				cfg.Password = "guest"
				cfg.VHost = "/"
				cfg.TLS.Enabled = false
				return cfg
			},
			want: "amqp://guest:guest@localhost:5672/",
		},
		{
			name: "connection with TLS",
			setupConfig: func() rabbitmq.ConnectionConfig {
				cfg := rabbitmq.DefaultConnectionConfig()
				cfg.Host = "rabbitmq.example.com"
				cfg.Port = 5671
				cfg.User = "admin"
				cfg.Password = "secret"
				cfg.VHost = "/production"
				cfg.TLS.Enabled = true
				return cfg
			},
			want: "amqps://admin:secret@rabbitmq.example.com:5671/production",
		},
		{
			name: "connection with custom vhost",
			setupConfig: func() rabbitmq.ConnectionConfig {
				cfg := rabbitmq.DefaultConnectionConfig()
				cfg.Host = "localhost"
				cfg.Port = 5672
				cfg.User = "user"
				cfg.Password = "pass"
				cfg.VHost = "/my-vhost"
				return cfg
			},
			want: "amqp://user:pass@localhost:5672/my-vhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			got := cfg.BuildURI()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueueConfig_WithDeadLetterExchange(t *testing.T) {
	//nolint:exhaustruct // test struct
	qc := rabbitmq.QueueConfig{
		Name: "test-queue",
		Args: nil,
	}
	result := qc.WithDeadLetterExchange("test-queue.dlx")

	assert.NotNil(t, result.Args)
	assert.Equal(t, "test-queue.dlx", result.Args["x-dead-letter-exchange"])
}

func TestQueueConfig_WithDeadLetterExchange_NilArgs(t *testing.T) {
	//nolint:exhaustruct // test struct
	qc := rabbitmq.QueueConfig{
		Name: "test-queue",
		Args: nil,
	}
	result := qc.WithDeadLetterExchange("test-queue.dlx")

	assert.NotNil(t, result.Args)
	assert.Equal(t, "test-queue.dlx", result.Args["x-dead-letter-exchange"])
}

func TestQueueConfig_WithDeliveryLimit(t *testing.T) {
	//nolint:exhaustruct // test struct
	qc := rabbitmq.QueueConfig{
		Name: "test-queue",
		Args: nil,
	}
	result := qc.WithDeliveryLimit(5)

	assert.Equal(t, 5, result.Args["x-delivery-limit"])
}

func TestQueueConfig_WithTTL(t *testing.T) {
	//nolint:exhaustruct // test struct
	qc := rabbitmq.QueueConfig{
		Name: "test-queue",
		Args: nil,
	}
	result := qc.WithTTL(30 * time.Second)

	assert.Equal(t, 30000, result.Args["x-message-ttl"])
}

func TestQueueConfig_WithMaxPriority(t *testing.T) {
	//nolint:exhaustruct // test struct
	qc := rabbitmq.QueueConfig{
		Name: "test-queue",
		Args: nil,
	}
	result := qc.WithMaxPriority(10)

	assert.Equal(t, 10, result.Args["x-max-priority"])
}

func TestNewExchangeConfig(t *testing.T) {
	cfg := rabbitmq.NewExchangeConfig("test.exchange", rabbitmq.ExchangeTypeTopic)

	assert.Equal(t, "test.exchange", cfg.Name)
	assert.Equal(t, rabbitmq.ExchangeTypeTopic, cfg.Type)
	assert.True(t, cfg.Durable)
	assert.False(t, cfg.AutoDelete)
	assert.False(t, cfg.Internal)
	assert.False(t, cfg.NoWait)
	assert.NotNil(t, cfg.Args)
}

func TestNewQueueConfig(t *testing.T) {
	cfg := rabbitmq.NewQueueConfig("test-queue", rabbitmq.QueueTypeQuorum)

	assert.Equal(t, "test-queue", cfg.Name)
	assert.Equal(t, rabbitmq.QueueTypeQuorum, cfg.Type)
	assert.True(t, cfg.Durable)
	assert.False(t, cfg.AutoDelete)
	assert.False(t, cfg.Exclusive)
	assert.False(t, cfg.NoWait)
	assert.NotNil(t, cfg.Args)
	assert.Equal(t, rabbitmq.QueueTypeQuorum, cfg.Args["x-queue-type"])
}

func TestNewBindingConfig(t *testing.T) {
	cfg := rabbitmq.NewBindingConfig("test-queue", "test.exchange", "routing.key")

	assert.Equal(t, "test-queue", cfg.QueueName)
	assert.Equal(t, "test.exchange", cfg.ExchangeName)
	assert.Equal(t, "routing.key", cfg.RoutingKey)
	assert.False(t, cfg.NoWait)
	assert.Nil(t, cfg.Args)
}
