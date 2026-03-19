package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/rabbitmq"
	"recipes-desk/pkg/pool"
)

// Publisher implements ports.EventBusWithConfirm for RabbitMQ.
type Publisher struct {
	pool    pool.Pool[*amqp.Channel]
	log     *zerolog.Logger
	options Options
}

// New creates a new RabbitMQ event publisher.
func New(
	ctx context.Context,
	pool pool.Pool[*amqp.Channel],
	log *zerolog.Logger,
	optionFuncs ...OptionFunc,
) (*Publisher, error) {
	options := getDefaultOptions()
	for _, optionFunc := range optionFuncs {
		optionFunc(&options)
	}

	publisher := &Publisher{
		pool:    pool,
		log:     log,
		options: options,
	}

	ch, err := pool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}
	defer pool.Put(ch)
	if err = rabbitmq.DeclareExchange(ch, options.Exchange); err != nil {
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return publisher, nil
}

// Publish publishes an event to RabbitMQ without waiting for confirmation.
func (p *Publisher) Publish(ctx context.Context, event rabbitmq.Event) error {
	ch, err := p.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}
	defer p.pool.Put(ch)

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err = ch.Publish(
		p.options.Exchange.Name,
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
		return fmt.Errorf("publish event: %w", err)
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
func (p *Publisher) PublishWithConfirm(ctx context.Context, event rabbitmq.Event) error {
	ch, err := p.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}
	defer p.pool.Put(ch)

	// Enable publisher confirms
	if err = ch.Confirm(false); err != nil {
		return fmt.Errorf("enable publisher confirms: %w", err)
	}

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err = ch.Publish(
		p.options.Exchange.Name,
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
		return fmt.Errorf("publish event: %w", err)
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
