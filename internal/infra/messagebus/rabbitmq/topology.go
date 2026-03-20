// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DeclareExchange declares an exchange with the given configuration.
func (c *Connection) DeclareExchange(cfg ExchangeConfig) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare(
		cfg.Name,
		string(cfg.Type),
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Internal,
		cfg.NoWait,
		cfg.Args,
	); err != nil {
		return fmt.Errorf("%w %s: %w", ErrDeclareExchange, cfg.Name, err)
	}

	return nil
}

// DeclareQueue declares a queue with the given configuration.
func (c *Connection) DeclareQueue(cfg QueueConfig) (amqp.Queue, error) {
	ch, err := c.Channel()
	if err != nil {
		return amqp.Queue{}, err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		cfg.Name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		cfg.Args,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("%w %s: %w", ErrDeclareQueue, cfg.Name, err)
	}

	return q, nil
}

// BindQueue binds a queue to an exchange.
func (c *Connection) BindQueue(cfg BindingConfig) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err = ch.QueueBind(
		cfg.QueueName,
		cfg.RoutingKey,
		cfg.ExchangeName,
		cfg.NoWait,
		cfg.Args,
	); err != nil {
		return fmt.Errorf(
			"%w %s to exchange %s: %w",
			ErrBindQueue,
			cfg.QueueName,
			cfg.ExchangeName,
			err,
		)
	}

	return nil
}

// DeleteExchange deletes an exchange.
func (c *Connection) DeleteExchange(name string, ifUnused bool) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err = ch.ExchangeDelete(name, ifUnused, false); err != nil {
		return fmt.Errorf("%w %s: %w", ErrDeclareExchange, name, err)
	}

	return nil
}

// DeleteQueue deletes a queue.
func (c *Connection) DeleteQueue(name string, ifUnused, ifEmpty bool) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err = ch.QueueDelete(name, ifUnused, ifEmpty, false); err != nil {
		return fmt.Errorf("%w %s: %w", ErrDeleteQueue, name, err)
	}

	return nil
}

// PurgeQueue removes all messages from a queue.
func (c *Connection) PurgeQueue(name string) (int, error) {
	ch, err := c.Channel()
	if err != nil {
		return 0, err
	}
	defer ch.Close()

	count, err := ch.QueuePurge(name, false)
	if err != nil {
		return 0, fmt.Errorf("%w %s: %w", ErrPurgeQueue, name, err)
	}

	return count, nil
}
