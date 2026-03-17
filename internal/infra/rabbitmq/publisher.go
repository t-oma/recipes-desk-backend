package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
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
	connManager  *ConnectionManager
	exchange     string
	exchangeType string
	log          *zerolog.Logger
	pool         *ChannelPool
	mu           sync.RWMutex
	closed       bool
}

// NewPublisher creates a new RabbitMQ event publisher.
func NewPublisher(
	ctx context.Context,
	connManager *ConnectionManager,
	exchange string,
	log *zerolog.Logger,
	poolSize int,
) (*Publisher, error) {
	pool, err := NewChannelPool(connManager, poolSize, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel pool: %w", err)
	}

	publisher := &Publisher{
		connManager:  connManager,
		exchange:     exchange,
		exchangeType: "topic",
		log:          log,
		pool:         pool,
		mu:           sync.RWMutex{},
		closed:       false,
	}

	if err = publisher.exchangeDeclare(ctx); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return publisher, nil
}

// Publish publishes an event to RabbitMQ without waiting for confirmation.
func (p *Publisher) Publish(ctx context.Context, event Event) error {
	ch, err := p.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer p.pool.Put(ch)

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

	if p.log != nil {
		p.log.Debug().
			Str("event_type", event.Type()).
			Str("routing_key", event.RoutingKey()).
			Msg("Event published")
	}

	return nil
}

// PublishWithConfirm publishes an event and waits for confirmation from RabbitMQ.
func (p *Publisher) PublishWithConfirm(ctx context.Context, event Event) error {
	ch, err := p.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer p.pool.Put(ch)

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
			if p.log != nil {
				p.log.Debug().
					Str("event_type", event.Type()).
					Str("routing_key", event.RoutingKey()).
					Msg("Event published with confirmation")
			}
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

// Close closes the publisher and releases all resources.
func (p *Publisher) Close() error {
	if p.IsClosed() {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	p.pool.Close()

	return nil
}

func (p *Publisher) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// exchangeDeclare declares the exchange on RabbitMQ.
func (p *Publisher) exchangeDeclare(ctx context.Context) error {
	ch, err := p.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer p.pool.Put(ch)

	if err = ch.ExchangeDeclare(
		p.exchange,
		p.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	return nil
}
