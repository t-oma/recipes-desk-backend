package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Event is the interface for all domain events.
type Event interface {
	Type() string
	RoutingKey() string
}

// Publisher implements ports.EventBusWithConfirm for RabbitMQ.
type Publisher struct {
	conn         *amqp.Connection
	exchange     string
	exchangeType string
	log          *zerolog.Logger
}

// NewPublisher creates a new RabbitMQ event publisher.
func NewPublisher(conn *amqp.Connection, exchange string, log *zerolog.Logger) *Publisher {
	return &Publisher{
		conn:         conn,
		exchange:     exchange,
		exchangeType: "topic",
		log:          log,
	}
}

// Publish publishes an event to RabbitMQ without waiting for confirmation.
func (p *Publisher) Publish(_ context.Context, event Event) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare(
		p.exchange,     // name
		p.exchangeType, // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // noWait
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err = ch.Publish(
		p.exchange,
		event.RoutingKey(),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
			Type:        event.Type(),
		},
	); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	p.log.Debug().
		Str("event_type", event.Type()).
		Str("routing_key", event.RoutingKey()).
		Msg("Event published")

	return nil
}

// PublishWithConfirm publishes an event and waits for confirmation from RabbitMQ.
func (p *Publisher) PublishWithConfirm(
	ctx context.Context,
	event Event,
) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare(
		p.exchange,     // name
		p.exchangeType, // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // noWait
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Enable publisher confirms
	if err = ch.Confirm(false); err != nil {
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err = ch.Publish(
		p.exchange,
		event.RoutingKey(),
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			Type:         event.Type(),
			DeliveryMode: amqp.Persistent,
		},
	); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Wait for confirmation
	select {
	case confirm := <-confirms:
		if confirm.Ack {
			p.log.Debug().
				Str("event_type", event.Type()).
				Str("routing_key", event.RoutingKey()).
				Msg("Event published with confirmation")
			return nil
		}
		return errors.New("event was nacked by RabbitMQ")

	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf(
				"context deadline exceeded while waiting for confirmation: %w",
				ctx.Err(),
			)
		}
		return fmt.Errorf("context cancelled while waiting for confirmation: %w", ctx.Err())
	}
}
