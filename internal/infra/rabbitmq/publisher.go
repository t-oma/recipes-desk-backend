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
	channelPool  chan *amqp.Channel
	poolSize     int
	mu           sync.RWMutex
	closed       bool
}

// NewPublisher creates a new RabbitMQ event publisher.
func NewPublisher(
	connManager *ConnectionManager,
	exchange string,
	log *zerolog.Logger,
	poolSize int,
) *Publisher {
	if poolSize <= 0 {
		poolSize = 10
	}

	return &Publisher{
		connManager:  connManager,
		exchange:     exchange,
		exchangeType: "topic",
		log:          log,
		channelPool:  make(chan *amqp.Channel, poolSize),
		poolSize:     poolSize,
		mu:           sync.RWMutex{},
		closed:       false,
	}
}

// Initialize creates the channel pool.
func (p *Publisher) Initialize() error {
	if p.IsClosed() {
		return errors.New("publisher is closed")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < p.poolSize; i++ {
		ch, err := p.createChannel()
		if err != nil {
			// Close already created channels
			p.closePool()
			return fmt.Errorf("failed to create channel %d: %w", i, err)
		}
		p.channelPool <- ch
	}

	return nil
}

// ExchangeDeclare declares the exchange on RabbitMQ.
func (p *Publisher) ExchangeDeclare() error {
	ch, err := p.getChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer p.returnChannel(ch)

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

// Publish publishes an event to RabbitMQ without waiting for confirmation.
func (p *Publisher) Publish(_ context.Context, event Event) error {
	ch, err := p.getChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer p.returnChannel(ch)

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
	ch, err := p.getChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer p.returnChannel(ch)

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
	close(p.channelPool)
	p.closePool()

	return nil
}

func (p *Publisher) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// getChannel retrieves a channel from the pool.
func (p *Publisher) getChannel() (*amqp.Channel, error) {
	if p.IsClosed() {
		return nil, errors.New("publisher is closed")
	}

	select {
	case ch := <-p.channelPool:
		// Check if channel is still valid
		if ch.IsClosed() {
			// Recreate channel
			newCh, err := p.createChannel()
			if err != nil {
				return nil, err
			}
			return newCh, nil
		}
		return ch, nil
	case <-time.After(5 * time.Second):
		return nil, errors.New("timeout waiting for channel from pool")
	}
}

// returnChannel returns a channel to the pool.
func (p *Publisher) returnChannel(ch *amqp.Channel) {
	if ch == nil || ch.IsClosed() {
		return
	}

	if p.IsClosed() {
		ch.Close()
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case p.channelPool <- ch:
	default:
		// Pool is full, close the channel
		ch.Close()
	}
}

// createChannel creates a new channel from the connection.
func (p *Publisher) createChannel() (*amqp.Channel, error) {
	conn, err := p.connManager.Connection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return ch, nil
}

// closePool closes all channels in the pool.
func (p *Publisher) closePool() {
	for {
		select {
		case ch := <-p.channelPool:
			if ch != nil {
				ch.Close()
			}
		default:
			return
		}
	}
}
