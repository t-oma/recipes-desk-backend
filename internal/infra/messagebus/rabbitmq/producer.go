// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/messagebus"
)

// Producer handles message publishing to RabbitMQ.
type Producer struct {
	conn          *Connection
	channel       *amqp.Channel
	confirms      chan amqp.Confirmation
	mu            sync.RWMutex
	log           *zerolog.Logger
	isClosed      bool
	isConfirmMode bool
}

// NewProducer creates a new message producer.
func NewProducer(conn *Connection, log *zerolog.Logger) (*Producer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	//nolint:exhaustruct // intentionally leaving optional fields empty
	p := &Producer{
		conn:          conn,
		channel:       ch,
		log:           log,
		confirms:      make(chan amqp.Confirmation, 1),
		isConfirmMode: false,
		isClosed:      false,
	}

	return p, nil
}

// NewProducerWithConfirms creates a new producer with publisher confirms enabled.
func NewProducerWithConfirms(conn *Connection, log *zerolog.Logger) (*Producer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Enable publisher confirms
	if err = ch.Confirm(false); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	//nolint:exhaustruct // intentionally leaving optional fields empty
	p := &Producer{
		conn:          conn,
		channel:       ch,
		log:           log,
		isConfirmMode: true,
		confirms:      ch.NotifyPublish(make(chan amqp.Confirmation, 1)),
		isClosed:      false,
	}

	return p, nil
}

// Publish sends a message to the specified exchange with the given routing key.
func (p *Producer) Publish(ctx context.Context, exchange, routingKey string,
	msg messagebus.Message,
) error {
	p.mu.RLock()
	if p.isClosed {
		p.mu.RUnlock()
		return ErrChannelClosed
	}
	ch := p.channel
	p.mu.RUnlock()

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
		select {
		case confirm := <-p.confirms:
			if !confirm.Ack {
				return ErrNackReceived
			}
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ErrPublishTimeout, ctx.Err())
		case <-time.After(20 * time.Second):
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

// Close closes the producer and its channel.
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed {
		return nil
	}

	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	p.isClosed = true
	p.log.Info().Msg("Producer closed")

	return nil
}

// IsClosed returns true if the producer is closed.
func (p *Producer) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isClosed
}
