package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// HandlerFunc is the function signature for event handlers.
type HandlerFunc func(ctx context.Context, routingKey string, body []byte) error

// Consumer handles consuming events from RabbitMQ with retry and DLQ.
type Consumer struct {
	config         ConsumerConfig
	pool           *ChannelPool
	log            *zerolog.Logger
	declaredQueues map[string]bool
}

// NewConsumer creates a new RabbitMQ consumer.
func NewConsumer(
	config ConsumerConfig,
	pool *ChannelPool,
	log *zerolog.Logger,
) *Consumer {
	config = config.WithDefaults()

	return &Consumer{
		config:         config,
		pool:           pool,
		log:            log,
		declaredQueues: make(map[string]bool),
	}
}

// Setup declares exchange, main queue, retry queues, and DLQ.
func (c *Consumer) Setup(ctx context.Context) error {
	ch, err := c.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare(
		c.config.Exchange, // name
		"topic",           // kind
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // noWait
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	dlqName, err := c.declareDLQ(ch)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Declare retry queues with TTL
	for retry := 1; retry <= c.config.MaxRetries; retry++ {
		if _, err = c.declareRetryQueue(ch, retry); err != nil {
			return fmt.Errorf("failed to declare retry queue %d: %w", retry, err)
		}
	}

	_, err = c.declareMainQueue(ch, dlqName)
	if err != nil {
		return fmt.Errorf("failed to declare main queue: %w", err)
	}

	// Bind main queue to exchange
	if err = ch.QueueBind(
		c.config.Queue,
		c.config.RoutingKey,
		c.config.Exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	c.log.Info().
		Str("exchange", c.config.Exchange).
		Str("queue", c.config.Queue).
		Str("routing_key", c.config.RoutingKey).
		Int("max_retries", c.config.MaxRetries).
		Msg("Consumer setup complete")

	return nil
}

// Start begins consuming messages.
func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			c.log.Info().Msg("Consumer stopped")
			return nil
		default:
		}

		if err := c.consumeLoop(ctx); err != nil {
			c.log.Error().Err(err).Msg("Consume loop error, restarting...")
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(1 * time.Second):
				continue
			}
		}
	}
}

func (c *Consumer) consumeLoop(ctx context.Context) error {
	ch, err := c.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	// Set QoS - prefetch 1 message at a time
	if err = ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(
		c.config.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.log.Info().Str("queue", c.config.Queue).Msg("Started consuming messages")

	for {
		select {
		case <-ctx.Done():
			c.log.Info().Msg("Consumer stopped")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				c.log.Warn().Msg("Message channel closed")
				return nil
			}

			if err = c.processMessage(ctx, ch, &msg); err != nil {
				c.log.Error().
					Err(err).
					Str("routing_key", msg.RoutingKey).
					Msg("Failed to process message")
			}
		}
	}
}

// declareDLQ declares the Dead Letter Queue (DLQ).
func (c *Consumer) declareDLQ(ch *amqp.Channel) (string, error) {
	dlqName := c.config.Queue + "-dlq"
	_, err := ch.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return "", err
	}
	c.declaredQueues[dlqName] = true

	return dlqName, nil
}

func (c *Consumer) declareRetryQueue(ch *amqp.Channel, retry int) (string, error) {
	retryQueueName := newRetryQueueName(c.config.Queue, retry)
	ttl := c.config.GetRetryTTL(retry)

	_, err := ch.QueueDeclare(
		retryQueueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "", // default exchange
			"x-dead-letter-routing-key": c.config.Queue,
			"x-message-ttl":             ttl.Milliseconds(),
		},
	)
	if err != nil {
		return "", err
	}
	c.declaredQueues[retryQueueName] = true

	return retryQueueName, nil
}

func (c *Consumer) declareMainQueue(ch *amqp.Channel, dlqName string) (string, error) {
	_, err := ch.QueueDeclare(
		c.config.Queue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "", // default exchange
			"x-dead-letter-routing-key": dlqName,
		},
	)
	if err != nil {
		return "", err
	}
	c.declaredQueues[c.config.Queue] = true

	return c.config.Queue, nil
}

func (c *Consumer) processMessage(ctx context.Context, ch *amqp.Channel, msg *amqp.Delivery) error {
	// Get retry count from headers
	retryCount := 0
	if msg.Headers != nil {
		if count, ok := msg.Headers["x-retry-count"].(int32); ok {
			retryCount = int(count)
		}
	}

	c.log.Debug().
		Str("routing_key", msg.RoutingKey).
		Int("retry_count", retryCount).
		Str("message_id", msg.MessageId).
		Msg("Processing message")

	if err := c.config.Handler(ctx, msg.RoutingKey, msg.Body); err != nil {
		c.log.Error().Err(err).Int("retry_count", retryCount).Msg("Handler failed")

		// Check if we should retry
		if retryCount < c.config.MaxRetries {
			return c.retryMessage(ctx, ch, msg, retryCount)
		}

		// Max retries reached - reject and send to DLQ
		c.log.Error().
			Int("retry_count", retryCount).
			Str("message_id", msg.MessageId).
			Msg("Max retries reached, sending to DLQ")

		if err = msg.Reject(false); err != nil {
			return fmt.Errorf("failed to reject message: %w", err)
		}

		return nil
	}

	// Success - acknowledge message
	if err := msg.Ack(false); err != nil {
		return fmt.Errorf("failed to acknowledge message: %w", err)
	}

	return nil
}

func (c *Consumer) retryMessage(
	_ context.Context,
	ch *amqp.Channel,
	msg *amqp.Delivery,
	retryCount int,
) error {
	retryQueue := newRetryQueueName(c.config.Queue, retryCount+1)

	// Publish to retry queue
	if err := ch.Publish(
		"",         // default exchange
		retryQueue, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType:   msg.ContentType,
			Body:          msg.Body,
			Headers:       amqp.Table{"x-retry-count": int32(retryCount + 1)},
			Timestamp:     msg.Timestamp,
			MessageId:     msg.MessageId,
			CorrelationId: msg.CorrelationId,
			Type:          msg.Type,
		},
	); err != nil {
		return fmt.Errorf("failed to publish to retry queue: %w", err)
	}

	// Acknowledge original message so it's removed from main queue
	if err := msg.Ack(false); err != nil {
		return fmt.Errorf("failed to acknowledge message after retry: %w", err)
	}

	c.log.Info().
		Int("retry_count", retryCount+1).
		Str("retry_queue", retryQueue).
		Str("message_id", msg.MessageId).
		Msg("Message sent to retry queue")

	return nil
}

func newRetryQueueName(queue string, retry int) string {
	return fmt.Sprintf("%s-retry-%d", queue, retry)
}
