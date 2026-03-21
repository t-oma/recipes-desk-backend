// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/messagebus"
)

// Consumer handles message consumption from RabbitMQ.
type Consumer struct {
	conn           *Connection
	log            *zerolog.Logger
	isClosed       atomic.Bool
	prefetch       int
	handlerTimeout time.Duration
	wg             sync.WaitGroup
}

// NewConsumer creates a new message consumer.
func NewConsumer(conn *Connection, log *zerolog.Logger, prefetch int) (*Consumer, error) {
	//nolint:exhaustruct // wg initialized automatically
	return &Consumer{
		conn:     conn,
		log:      log,
		prefetch: prefetch,
	}, nil
}

// SetHandlerTimeout sets the timeout for message handler execution.
func (c *Consumer) SetHandlerTimeout(timeout time.Duration) {
	c.handlerTimeout = timeout
}

// IsClosed returns true if the consumer is closed.
func (c *Consumer) IsClosed() bool {
	return c.isClosed.Load()
}

// Close closes the consumer and waits for in-flight messages to complete.
func (c *Consumer) Close() error {
	if !c.isClosed.CompareAndSwap(false, true) {
		return nil
	}

	c.log.Info().Msg("Consumer closing, waiting for in-flight messages...")
	c.wg.Wait()
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
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	defer ch.Close()

	if c.prefetch > 0 {
		if qosErr := ch.Qos(c.prefetch, 0, false); qosErr != nil {
			return fmt.Errorf("failed to set prefetch: %w", qosErr)
		}
	}

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
			c.wg.Add(1)
			go c.handleMessage(ctx, msg, handler)
		}
	}
}

// handleMessage processes a single message.
func (c *Consumer) handleMessage(
	ctx context.Context,
	delivery amqp.Delivery,
	handler messagebus.Handler,
) {
	defer c.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			c.log.Error().
				Any("panic", r).
				Str("message_id", delivery.MessageId).
				Msg("Panic in message handler")
			if err := delivery.Nack(false, false); err != nil {
				c.log.Error().Err(err).Msg("Failed to nack message")
			}
		}
	}()

	var message messagebus.Message
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		c.log.Error().
			Err(err).
			Str("message_id", delivery.MessageId).
			Msg("Failed to unmarshal message")
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

	handlerCtx := ctx
	if c.handlerTimeout > 0 {
		var cancel context.CancelFunc
		handlerCtx, cancel = context.WithTimeout(ctx, c.handlerTimeout)
		defer cancel()
	}

	if err := handler(handlerCtx, message); err != nil {
		c.log.Error().
			Err(err).
			Str("message_id", message.ID).
			Str("message_type", message.Type).
			Msg("Handler failed, sending to DLQ")
		if err = delivery.Nack(false, false); err != nil {
			c.log.Error().Err(err).Msg("Failed to nack message")
		}
		return
	}

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
