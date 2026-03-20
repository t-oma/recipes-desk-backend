// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/messagebus"
)

// Consumer handles message consumption from RabbitMQ.
type Consumer struct {
	conn       *Connection
	channel    *amqp.Channel
	mu         sync.RWMutex
	log        *zerolog.Logger
	isClosed   bool
	prefetch   int
	autoAck    bool
	handlers   map[string]messagebus.Handler
	handlersMu sync.RWMutex
}

// NewConsumer creates a new message consumer.
func NewConsumer(conn *Connection, log *zerolog.Logger, prefetch int) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Set prefetch
	if prefetch > 0 {
		if err = ch.Qos(prefetch, 0, false); err != nil {
			ch.Close()
			return nil, fmt.Errorf("failed to set prefetch: %w", err)
		}
	}

	//nolint:exhaustruct // intentionally leaving optional fields empty
	c := &Consumer{
		conn:     conn,
		channel:  ch,
		log:      log,
		prefetch: prefetch,
		autoAck:  false, // Manual ACK for reliability
		handlers: make(map[string]messagebus.Handler),
	}

	return c, nil
}

// Consume starts consuming messages from the specified queue.
func (c *Consumer) Consume(ctx context.Context, queue string, handler messagebus.Handler) error {
	c.mu.RLock()
	if c.isClosed {
		c.mu.RUnlock()
		return ErrChannelClosed
	}
	ch := c.channel
	c.mu.RUnlock()

	c.handlersMu.Lock()
	c.handlers[queue] = handler
	c.handlersMu.Unlock()

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

	go c.processMessages(ctx, msgs, handler)

	return nil
}

// DeclareQueue declares a queue with the given configuration.
func (c *Consumer) DeclareQueue(cfg QueueConfig) error {
	c.mu.RLock()
	if c.isClosed {
		c.mu.RUnlock()
		return ErrChannelClosed
	}
	ch := c.channel
	c.mu.RUnlock()

	_, err := ch.QueueDeclare(
		cfg.Name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		cfg.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", cfg.Name, err)
	}

	c.log.Debug().
		Str("queue", cfg.Name).
		Str("type", string(cfg.Type)).
		Bool("durable", cfg.Durable).
		Msg("Queue declared")

	return nil
}

// BindQueue binds a queue to an exchange.
func (c *Consumer) BindQueue(cfg BindingConfig) error {
	c.mu.RLock()
	if c.isClosed {
		c.mu.RUnlock()
		return ErrChannelClosed
	}
	ch := c.channel
	c.mu.RUnlock()

	if err := ch.QueueBind(
		cfg.QueueName,
		cfg.RoutingKey,
		cfg.ExchangeName,
		cfg.NoWait,
		cfg.Args,
	); err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w",
			cfg.QueueName, cfg.ExchangeName, err)
	}

	c.log.Debug().
		Str("queue", cfg.QueueName).
		Str("exchange", cfg.ExchangeName).
		Str("routing_key", cfg.RoutingKey).
		Msg("Queue bound")

	return nil
}

// Close closes the consumer and its channel.
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return nil
	}

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	c.isClosed = true
	c.log.Info().Msg("Consumer closed")

	return nil
}

// IsClosed returns true if the consumer is closed.
func (c *Consumer) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isClosed
}

// processMessages processes incoming messages.
func (c *Consumer) processMessages(
	ctx context.Context,
	msgs <-chan amqp.Delivery,
	handler messagebus.Handler,
) {
	for {
		select {
		case <-ctx.Done():
			c.log.Info().Msg("Consumer stopped due to context cancellation")
			return
		case msg, ok := <-msgs:
			if !ok {
				c.log.Info().Msg("Message channel closed")
				return
			}

			c.handleMessage(ctx, msg, handler)
		}
	}
}

// handleMessage processes a single message.
func (c *Consumer) handleMessage(
	ctx context.Context,
	msg amqp.Delivery,
	handler messagebus.Handler,
) {
	var message messagebus.Message
	if err := json.Unmarshal(msg.Body, &message); err != nil {
		c.log.Error().
			Err(err).
			Str("message_id", msg.MessageId).
			Msg("Failed to unmarshal message")
		// NACK without requeue - send to DLQ
		if err = msg.Nack(false, false); err != nil {
			c.log.Error().Err(err).Msg("Failed to nack message")
		}
		return
	}

	c.log.Debug().
		Str("message_id", message.ID).
		Str("message_type", message.Type).
		Str("routing_key", msg.RoutingKey).
		Msg("Received message")

	if err := handler(ctx, message); err != nil {
		c.log.Error().
			Err(err).
			Str("message_id", message.ID).
			Str("message_type", message.Type).
			Msg("Handler failed")
		// NACK with requeue for retry
		if err = msg.Nack(false, true); err != nil {
			c.log.Error().Err(err).Msg("Failed to nack message")
		}
		return
	}

	// ACK on success
	if err := msg.Ack(false); err != nil {
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
