// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"

	gonanoid "github.com/matoous/go-nanoid/v2"
	amqp "github.com/rabbitmq/amqp091-go"

	"recipes-desk/internal/infra/messagebus"
)

// Producer handles message publishing to RabbitMQ.
type Producer struct {
	conn     *Connection
	log      *zerolog.Logger
	config   ProducerConfig
	isClosed atomic.Bool
}

// NewProducer creates a new message producer.
func NewProducer(conn *Connection, log *zerolog.Logger, config ProducerConfig) (*Producer, error) {
	//nolint:exhaustruct // isClosed initialized automatically
	return &Producer{
		conn:   conn,
		log:    log,
		config: config,
	}, nil
}

// Close closes the producer.
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
	if err := ensureMessage(&msg); err != nil {
		return err
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	defer ch.Close()

	if p.config.ConfirmMode {
		if err = ch.Confirm(false); err != nil {
			return fmt.Errorf("failed to enable publisher confirms: %w", err)
		}
	}

	publishing, err := buildPublishing(msg)
	if err != nil {
		return fmt.Errorf("failed to build publishing: %w", err)
	}

	confirms, returns := p.registerNotifyChannels(ch)

	p.log.Debug().
		Str("exchange", exchange).
		Str("routing_key", routingKey).
		Str("message_id", msg.ID).
		Str("message_type", msg.Type).
		Msg("Publishing message")

	if err = ch.Publish(exchange, routingKey, p.config.Mandatory, false, publishing); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	if err = p.waitForResult(ctx, confirms, returns); err != nil {
		return err
	}

	p.log.Info().
		Str("exchange", exchange).
		Str("routing_key", routingKey).
		Str("message_id", msg.ID).
		Msg("Message published successfully")

	return nil
}

func (p *Producer) registerNotifyChannels(
	ch *amqp.Channel,
) (chan amqp.Confirmation, chan amqp.Return) {
	var confirms chan amqp.Confirmation
	var returns chan amqp.Return

	if p.config.ConfirmMode {
		confirms = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	}

	if p.config.Mandatory {
		returns = ch.NotifyReturn(make(chan amqp.Return, 1))
	}

	return confirms, returns
}

func (p *Producer) waitForResult(
	ctx context.Context,
	confirms chan amqp.Confirmation,
	returns chan amqp.Return,
) error {
	switch {
	case p.config.ConfirmMode && p.config.Mandatory:
		return p.waitForConfirmOrReturn(ctx, confirms, returns)
	case p.config.ConfirmMode:
		return p.waitForConfirm(ctx, confirms)
	case p.config.Mandatory:
		return p.waitForReturn(ctx, returns)
	default:
		return nil
	}
}

func (p *Producer) waitForConfirmOrReturn(
	ctx context.Context,
	confirms chan amqp.Confirmation,
	returns chan amqp.Return,
) error {
	select {
	case ret := <-returns:
		return fmt.Errorf(
			"%w: exchange=%s, routing_key=%s, reply_code=%d, reply_text=%s",
			ErrMandatoryFailed,
			ret.Exchange,
			ret.RoutingKey,
			ret.ReplyCode,
			ret.ReplyText,
		)
	case confirm := <-confirms:
		if !confirm.Ack {
			return ErrNackReceived
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.config.ConfirmTimeout):
		return ErrPublishTimeout
	}
}

func (p *Producer) waitForConfirm(ctx context.Context, confirms chan amqp.Confirmation) error {
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return ErrNackReceived
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.config.ConfirmTimeout):
		return ErrPublishTimeout
	}
}

func (p *Producer) waitForReturn(ctx context.Context, returns chan amqp.Return) error {
	select {
	case ret := <-returns:
		return fmt.Errorf("%w: exchange=%s, routing_key=%s, reply_code=%d, reply_text=%s",
			ErrMandatoryFailed, ret.Exchange, ret.RoutingKey, ret.ReplyCode, ret.ReplyText)
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.config.ConfirmTimeout):
		return ErrPublishTimeout
	}
}

func buildPublishing(msg messagebus.Message) (amqp.Publishing, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return amqp.Publishing{}, fmt.Errorf("failed to marshal message: %w", err)
	}

	headers := amqp.Table{}
	for k, v := range msg.Headers {
		headers[k] = v
	}

	//nolint:exhaustruct // intentionally leaving optional fields empty
	return amqp.Publishing{
		ContentType:  "application/json",
		MessageId:    msg.ID,
		Timestamp:    msg.Timestamp,
		Type:         msg.Type,
		Body:         body,
		Headers:      headers,
		DeliveryMode: amqp.Persistent,
	}, nil
}

func ensureMessage(msg *messagebus.Message) error {
	if msg.ID == "" {
		id, err := gonanoid.New()
		if err != nil {
			return fmt.Errorf("failed to generate message ID: %w", err)
		}
		msg.ID = id
	}

	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}

	return nil
}
