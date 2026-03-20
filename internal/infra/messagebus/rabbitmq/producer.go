// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/messagebus"
)

// Producer handles message publishing to RabbitMQ.
type Producer struct {
	conn           *Connection
	log            *zerolog.Logger
	isConfirmMode  bool
	isClosed       atomic.Bool
	confirmTimeout time.Duration
}

// NewProducer creates a new message producer.
func NewProducer(
	conn *Connection,
	log *zerolog.Logger,
	isConfirmMode bool,
) (*Producer, error) {
	return &Producer{
		conn:           conn,
		log:            log,
		isConfirmMode:  isConfirmMode,
		isClosed:       atomic.Bool{},
		confirmTimeout: 10 * time.Second,
	}, nil
}

func (p *Producer) SetConfirmTimeout(timeout time.Duration) {
	p.confirmTimeout = timeout
}

func (p *Producer) Close() error {
	if p.isClosed.Load() {
		return nil
	}

	p.isClosed.Store(true)
	p.log.Info().Msg("Producer closed")

	return nil
}

// Publish sends a message to the specified exchange with the given routing key.
func (p *Producer) Publish(ctx context.Context, exchange, routingKey string,
	msg messagebus.Message,
) error {
	// Create new channel for each publish (auto-recovery on reconnect)
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	defer ch.Close()

	if p.isConfirmMode {
		if err = ch.Confirm(false); err != nil {
			return fmt.Errorf("failed to enable publisher confirms: %w", err)
		}
	}

	// Marshal full message
	body, err := json.Marshal(messagebus.Message{
		ID:        msg.ID,
		Type:      msg.Type,
		Timestamp: msg.Timestamp,
		Source:    msg.Source,
		Payload:   msg.Payload,
		Headers:   msg.Headers,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Build AMQP headers
	headers := amqp.Table{}
	for k, v := range msg.Headers {
		headers[k] = v
	}

	//nolint:exhaustruct // intentionally leaving optional fields empty
	publishing := amqp.Publishing{
		ContentType:  "application/json",
		MessageId:    msg.ID,
		Timestamp:    msg.Timestamp,
		Type:         msg.Type,
		Body:         body,
		Headers:      headers,
		DeliveryMode: amqp.Persistent,
	}

	p.log.Debug().
		Str("exchange", exchange).
		Str("routing_key", routingKey).
		Str("message_id", msg.ID).
		Str("message_type", msg.Type).
		Msg("Publishing message")

	if err = ch.Publish(
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		publishing,
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation if in confirm mode
	if p.isConfirmMode {
		confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
		select {
		case confirm := <-confirms:
			if !confirm.Ack {
				return ErrNackReceived
			}
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ErrPublishTimeout, ctx.Err())
		case <-time.After(p.confirmTimeout):
			return ErrPublishTimeout
		}
	}

	p.log.Info().
		Str("exchange", exchange).
		Str("routing_key", routingKey).
		Str("message_id", msg.ID).
		Msg("Message published successfully")

	return nil
}
