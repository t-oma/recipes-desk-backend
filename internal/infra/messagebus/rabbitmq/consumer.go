// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/messagebus"
)

// Consumer handles message consumption from RabbitMQ.
type Consumer struct {
	conn     *Connection
	log      *zerolog.Logger
	isClosed atomic.Bool
	prefetch int
}

// NewConsumer creates a new message consumer.
func NewConsumer(conn *Connection, log *zerolog.Logger, prefetch int) (*Consumer, error) {
	//nolint:exhaustruct // mu initialized automatically
	return &Consumer{
		conn:     conn,
		log:      log,
		prefetch: prefetch,
	}, nil
}

// IsClosed returns true if the consumer is closed.
func (c *Consumer) IsClosed() bool {
	return c.isClosed.Load()
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	if c.isClosed.Load() {
		return nil
	}

	c.isClosed.Store(true)
	c.log.Info().Msg("Consumer closed")

	return nil
}

// Consume starts consuming messages from the specified queue with auto-recovery.
func (c *Consumer) Consume(ctx context.Context, queue string, handler messagebus.Handler) error {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			c.log.Info().Str("queue", queue).Msg("Consumer stopped")
			return ctx.Err()
		default:
		}

		if c.isClosed.Load() {
			c.log.Info().Str("queue", queue).Msg("Consumer closed")
			return nil
		}

		// Try to consume
		err := c.doConsume(ctx, queue, handler)
		if err != nil {
			c.log.Error().
				Err(err).
				Str("queue", queue).
				Msg("Consume failed, will retry")
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			c.log.Info().Str("queue", queue).Msg("Consumer stopped during backoff")
		case <-time.After(backoff):
			backoff = time.Duration(float64(backoff) * 1.5)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// doConsume creates channel, sets QoS, and starts consuming.
func (c *Consumer) doConsume(ctx context.Context, queue string, handler messagebus.Handler) error {
	// Create new channel
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	defer ch.Close()

	// Set QoS
	if c.prefetch > 0 {
		if qosErr := ch.Qos(c.prefetch, 0, false); qosErr != nil {
			return fmt.Errorf("failed to set prefetch: %w", qosErr)
		}
	}

	// Declare queue if not exists
	_, err = ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Start consuming
	msgs, err := ch.ConsumeWithContext(
		ctx,
		queue,
		"",    // consumer tag (auto-generated)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.log.Info().
		Str("queue", queue).
		Int("prefetch", c.prefetch).
		Msg("Started consuming messages")

	// Process messages
	for {
		select {
		case <-ctx.Done():
			c.log.Info().Str("queue", queue).Msg("Context cancelled, stopping consumer")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				c.log.Info().Str("queue", queue).Msg("Message channel closed, will reconnect")
				return errors.New("message channel closed")
			}
			c.handleMessage(ctx, msg, handler)
		}
	}
}

// handleMessage processes a single message.
func (c *Consumer) handleMessage(
	ctx context.Context,
	delivery amqp.Delivery,
	handler messagebus.Handler,
) {
	var message messagebus.Message
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		c.log.Error().
			Err(err).
			Str("message_id", delivery.MessageId).
			Msg("Failed to unmarshal message")
		// NACK without requeue - send to DLQ
		if err = delivery.Nack(false, false); err != nil {
			c.log.Error().Err(err).Msg("Failed to nack message")
		}
		return
	}

	c.log.Debug().
		Str("message_id", message.ID).
		Str("message_type", message.Type).
		Str("routing_key", delivery.RoutingKey).
		Msg("Received message")

	if err := handler(ctx, message); err != nil {
		c.log.Error().
			Err(err).
			Str("message_id", message.ID).
			Str("message_type", message.Type).
			Msg("Handler failed")
		// NACK with requeue for retry
		if err = delivery.Nack(false, true); err != nil {
			c.log.Error().Err(err).Msg("Failed to nack message")
		}
		return
	}

	// ACK on success
	if err := delivery.Ack(false); err != nil {
		c.log.Error().
			Err(err).
			Str("message_id", message.ID).
			Msg("Failed to ack message")
		return
	}

	c.log.Debug().
		Str("message_id", message.ID).
		Msg("Message processed successfully")
}
