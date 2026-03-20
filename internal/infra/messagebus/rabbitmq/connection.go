// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connection manages RabbitMQ connection with automatic reconnection.
type Connection struct {
	config       Config
	conn         *amqp.Connection
	log          *zerolog.Logger
	isConnected  atomic.Bool
	isClosed     atomic.Bool
	mu           sync.RWMutex
	closeChan    chan struct{}
	reconnecting atomic.Bool
}

// NewConnection creates a new RabbitMQ connection manager.
func NewConnection(cfg Config, log *zerolog.Logger) (*Connection, error) {
	//nolint:exhaustruct // fields initialized after connection
	c := &Connection{
		config:    cfg,
		log:       log,
		closeChan: make(chan struct{}),
	}

	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConnection, err)
	}

	return c, nil
}

// IsConnected returns true if the connection is active.
func (c *Connection) IsConnected() bool {
	return c.isConnected.Load()
}

// Channel creates a new channel on the connection.
func (c *Connection) Channel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil || !c.isConnected.Load() {
		return nil, ErrNotConnected
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConnection, err)
	}

	return ch, nil
}

// Close gracefully closes the connection.
func (c *Connection) Close() error {
	if c.isClosed.CompareAndSwap(false, true) {
		close(c.closeChan)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("%w: %w", ErrConnection, err)
		}
		c.conn = nil
		c.isConnected.Store(false)
	}

	c.log.Info().Msg("RabbitMQ connection closed")

	return nil
}

// Ping checks if the connection is healthy.
func (c *Connection) Ping(_ context.Context) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil || conn.IsClosed() {
		return ErrConnectionClosed
	}

	// Try to create a temporary channel to verify connection
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return nil
}

// connect establishes a connection to RabbitMQ.
func (c *Connection) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed.Load() {
		return ErrConnectionClosed
	}

	uri := c.config.buildURI()
	cfg := amqp.Config{
		Heartbeat: c.config.Heartbeat,
		Locale:    "en_US",
	}

	if c.config.ChannelMax > 0 {
		cfg.ChannelMax = c.config.ChannelMax
	}

	if c.config.TLS.Enabled {
		tlsConfig, err := c.buildTLSConfig()
		if err != nil {
			return fmt.Errorf("build TLS config: %w", err)
		}
		cfg.TLSClientConfig = tlsConfig
	}

	conn, err := amqp.DialConfig(uri, cfg)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConnection, err)
	}

	c.conn = conn
	c.isConnected.Store(true)

	c.log.Info().
		Str("host", c.config.Host).
		Int("port", c.config.Port).
		Str("vhost", c.config.VHost).
		Msg("Connected to RabbitMQ")

	// Start connection monitoring
	go c.handleReconnect()

	return nil
}

// buildTLSConfig creates TLS configuration.
func (c *Connection) buildTLSConfig() (*tls.Config, error) {
	if !c.config.TLS.Enabled {
		return nil, ErrNotConnected
	}

	//nolint:exhaustruct // intentionally using minimal TLS config
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: c.config.TLS.SkipVerify,
	}

	// TODO: Add certificate loading when files are provided
	// This is a placeholder for future TLS implementation

	return tlsConfig, nil
}

// handleReconnect monitors connection and handles reconnection.
func (c *Connection) handleReconnect() {
	for {
		c.mu.RLock()
		if c.conn == nil {
			c.mu.RUnlock()
			return
		}
		notifyClose := c.conn.NotifyClose(make(chan *amqp.Error, 1))
		c.mu.RUnlock()

		select {
		case <-c.closeChan:
			return
		case err := <-notifyClose:
			if err != nil {
				c.log.Error().
					Err(err).
					Msg("RabbitMQ connection closed unexpectedly")
			}

			c.isConnected.Store(false)

			if c.config.Reconnect.Enabled && !c.isClosed.Load() {
				c.attemptReconnect()
			}
			return
		}
	}
}

// attemptReconnect tries to reconnect with exponential backoff.
func (c *Connection) attemptReconnect() {
	if !c.reconnecting.CompareAndSwap(false, true) {
		return // Already reconnecting
	}
	defer c.reconnecting.Store(false)

	wait := c.config.Reconnect.InitialWait

	for attempt := 1; attempt <= c.config.Reconnect.MaxAttempts; attempt++ {
		if c.isClosed.Load() {
			return
		}

		c.log.Info().
			Int("attempt", attempt).
			Int("max_attempts", c.config.Reconnect.MaxAttempts).
			Dur("wait", wait).
			Msg("Attempting to reconnect to RabbitMQ")

		err := c.connect()
		if err == nil {
			c.log.Info().Msg("Successfully reconnected to RabbitMQ")
			return
		}

		c.log.Error().
			Err(err).
			Int("attempt", attempt).
			Msg("Failed to reconnect")

		if attempt < c.config.Reconnect.MaxAttempts {
			select {
			case <-c.closeChan:
				return
			case <-time.After(wait):
				wait = time.Duration(float64(wait) * c.config.Reconnect.Multiplier)
				if wait > c.config.Reconnect.MaxWait {
					wait = c.config.Reconnect.MaxWait
				}
			}
		}
	}

	c.log.Error().
		Int("max_attempts", c.config.Reconnect.MaxAttempts).
		Msg("Failed to reconnect to RabbitMQ after maximum attempts")
}
