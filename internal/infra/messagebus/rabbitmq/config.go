// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"fmt"
	"time"
)

// ExchangeType represents the type of RabbitMQ exchange.
type ExchangeType string

const (
	// ExchangeTypeDirect routes messages based on exact routing key match.
	ExchangeTypeDirect ExchangeType = "direct"
	// ExchangeTypeFanout broadcasts messages to all bound queues.
	ExchangeTypeFanout ExchangeType = "fanout"
	// ExchangeTypeTopic routes messages based on pattern matching.
	ExchangeTypeTopic ExchangeType = "topic"
	// ExchangeTypeHeaders routes messages based on header attributes.
	ExchangeTypeHeaders ExchangeType = "headers"
)

// QueueType represents the type of RabbitMQ queue.
type QueueType string

const (
	// QueueTypeClassic is the standard queue type.
	QueueTypeClassic QueueType = "classic"
	// QueueTypeQuorum is a replicated queue type for high availability.
	QueueTypeQuorum QueueType = "quorum"
)

// ConnectionConfig holds RabbitMQ connection configuration.
type ConnectionConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	VHost             string
	TLS               TLSConfig
	Reconnect         ReconnectConfig `yaml:"reconnect"`
	ChannelMax        uint16          `yaml:"channel-max"` // Maximum number of channels per connection
	Heartbeat         time.Duration   `yaml:"heartbeat"`   // Heartbeat interval
	ConnectionTimeout time.Duration   `yaml:"connection-timeout"`
}

// TLSConfig holds TLS configuration for RabbitMQ connection.
type TLSConfig struct {
	Enabled    bool
	CertFile   string
	KeyFile    string
	CAFile     string
	SkipVerify bool // Only for development
}

// ReconnectConfig holds reconnection configuration.
type ReconnectConfig struct {
	Enabled     bool          `yaml:"enabled"`
	MaxAttempts int           `yaml:"max-attempts"`
	InitialWait time.Duration `yaml:"initial-wait"`
	MaxWait     time.Duration `yaml:"max-wait"`
	Multiplier  float64       `yaml:"multiplier"`
}

// ExchangeConfig holds exchange declaration configuration.
type ExchangeConfig struct {
	Name       string
	Type       ExchangeType
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]interface{}
}

// QueueConfig holds queue declaration configuration.
type QueueConfig struct {
	Name       string
	Type       QueueType
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]any
}

// BindingConfig holds queue binding configuration.
type BindingConfig struct {
	QueueName    string
	ExchangeName string
	RoutingKey   string
	NoWait       bool
	Args         map[string]any
}

// DefaultConnectionConfig returns a default configuration.
func DefaultConnectionConfig() ConnectionConfig {
	//nolint:exhaustruct // intentionally using default values for optional fields
	return ConnectionConfig{
		Host:              "localhost",
		Port:              5672,
		VHost:             "/",
		ChannelMax:        2048,
		Heartbeat:         10 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		//nolint:exhaustruct // TLS disabled by default
		TLS: TLSConfig{
			Enabled:    false,
			SkipVerify: false,
		},
		Reconnect: ReconnectConfig{
			Enabled:     true,
			MaxAttempts: 10,
			InitialWait: 1 * time.Second,
			MaxWait:     30 * time.Second,
			Multiplier:  2,
		},
	}
}

// NewExchangeConfig creates a new exchange configuration with sensible defaults.
func NewExchangeConfig(name string, exchangeType ExchangeType) ExchangeConfig {
	return ExchangeConfig{
		Name:       name,
		Type:       exchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       make(map[string]interface{}),
	}
}

// NewQueueConfig creates a new queue configuration with sensible defaults.
func NewQueueConfig(name string, queueType QueueType) QueueConfig {
	return QueueConfig{
		Name:       name,
		Type:       queueType,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: map[string]any{
			"x-queue-type": queueType,
		},
	}
}

// NewBindingConfig creates a new binding configuration with sensible defaults.
func NewBindingConfig(queueName, exchangeName, routingKey string) BindingConfig {
	return BindingConfig{
		QueueName:    queueName,
		ExchangeName: exchangeName,
		RoutingKey:   routingKey,
		NoWait:       false,
		Args:         nil,
	}
}

// WithDeadLetterExchange adds dead letter exchange configuration to queue args.
func (qc QueueConfig) WithDeadLetterExchange(exchangeName string) QueueConfig {
	if qc.Args == nil {
		qc.Args = make(map[string]any)
	}
	qc.Args["x-dead-letter-exchange"] = exchangeName
	return qc
}

// WithDeliveryLimit sets the maximum delivery attempts for quorum queues.
func (qc QueueConfig) WithDeliveryLimit(limit int) QueueConfig {
	if qc.Args == nil {
		qc.Args = make(map[string]any)
	}
	qc.Args["x-delivery-limit"] = limit
	return qc
}

// WithTTL sets the time-to-live for messages in the queue.
func (qc QueueConfig) WithTTL(ttl time.Duration) QueueConfig {
	if qc.Args == nil {
		qc.Args = make(map[string]any)
	}
	qc.Args["x-message-ttl"] = int(ttl.Milliseconds())
	return qc
}

// WithMaxPriority sets the maximum priority for the queue.
func (qc QueueConfig) WithMaxPriority(priority int) QueueConfig {
	if qc.Args == nil {
		qc.Args = make(map[string]any)
	}
	qc.Args["x-max-priority"] = priority
	return qc
}

// ProducerConfig holds producer configuration.
type ProducerConfig struct {
	// ConfirmMode enables publisher confirms for reliable publishing.
	ConfirmMode bool
	// Mandatory flag - if true and no queue is bound to the routing key,
	// the message is returned to the publisher instead of being silently dropped.
	Mandatory bool
	// ConfirmTimeout is the timeout to wait for publisher confirmation.
	ConfirmTimeout time.Duration
}

// DefaultProducerConfig returns a default producer configuration.
func DefaultProducerConfig() ProducerConfig {
	return ProducerConfig{
		ConfirmMode:    true,
		Mandatory:      true,
		ConfirmTimeout: 5 * time.Second,
	}
}

// buildURI constructs the AMQP connection URI.
func (c ConnectionConfig) buildURI() string {
	scheme := "amqp"
	if c.TLS.Enabled {
		scheme = "amqps"
	}

	return fmt.Sprintf("%s://%s:%s@%s:%d%s",
		scheme,
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.VHost,
	)
}
