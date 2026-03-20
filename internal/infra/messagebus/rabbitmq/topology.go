// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DeclareExchange declares an exchange with the given configuration.
func DeclareExchange(conn *Connection, cfg ExchangeConfig) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
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
		return fmt.Errorf("declare exchange %s: %w", cfg.Name, err)
	}

	return nil
}

// DeclareQueue declares a queue with the given configuration.
func DeclareQueue(conn *Connection, cfg QueueConfig) (amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("create channel: %w", err)
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
		return amqp.Queue{}, fmt.Errorf("declare queue %s: %w", cfg.Name, err)
	}

	return q, nil
}

// BindQueue binds a queue to an exchange.
func BindQueue(conn *Connection, cfg BindingConfig) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	defer ch.Close()

	if err = ch.QueueBind(
		cfg.QueueName,
		cfg.RoutingKey,
		cfg.ExchangeName,
		cfg.NoWait,
		cfg.Args,
	); err != nil {
		return fmt.Errorf("bind queue %s to exchange %s: %w",
			cfg.QueueName, cfg.ExchangeName, err)
	}

	return nil
}

// DeleteExchange deletes an exchange.
func DeleteExchange(conn *Connection, name string, ifUnused bool) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	defer ch.Close()

	if err = ch.ExchangeDelete(name, ifUnused, false); err != nil {
		return fmt.Errorf("delete exchange %s: %w", name, err)
	}

	return nil
}

// DeleteQueue deletes a queue.
func DeleteQueue(conn *Connection, name string, ifUnused, ifEmpty bool) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	defer ch.Close()

	if _, err = ch.QueueDelete(name, ifUnused, ifEmpty, false); err != nil {
		return fmt.Errorf("delete queue %s: %w", name, err)
	}

	return nil
}

// PurgeQueue removes all messages from a queue.
func PurgeQueue(conn *Connection, name string) (int, error) {
	ch, err := conn.Channel()
	if err != nil {
		return 0, fmt.Errorf("create channel: %w", err)
	}
	defer ch.Close()

	count, err := ch.QueuePurge(name, false)
	if err != nil {
		return 0, fmt.Errorf("purge queue %s: %w", name, err)
	}

	return count, nil
}
