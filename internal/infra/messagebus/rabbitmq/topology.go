// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"fmt"

	"github.com/rs/zerolog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	dlxSuffix = ".dlx"
	dlqSuffix = ".dlq"
)

// Topology handles RabbitMQ topology operations.
type Topology struct {
	conn *Connection
	log  *zerolog.Logger
}

// NewTopology creates a new topology manager.
func NewTopology(conn *Connection, log *zerolog.Logger) *Topology {
	return &Topology{
		conn: conn,
		log:  log,
	}
}

// DeclareExchange declares an exchange with the given configuration.
func (t *Topology) DeclareExchange(cfg ExchangeConfig) error {
	ch, err := t.conn.Channel()
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

	t.log.Debug().
		Str("exchange", cfg.Name).
		Str("type", string(cfg.Type)).
		Bool("durable", cfg.Durable).
		Msg("Exchange declared")

	return nil
}

// DeclareQueue declares a queue with the given configuration.
func (t *Topology) DeclareQueue(cfg QueueConfig) (amqp.Queue, error) {
	ch, err := t.conn.Channel()
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

	t.log.Debug().
		Str("queue", cfg.Name).
		Str("type", string(cfg.Type)).
		Bool("durable", cfg.Durable).
		Msg("Queue declared")

	return q, nil
}

// BindQueue binds a queue to an exchange.
func (t *Topology) BindQueue(cfg BindingConfig) error {
	ch, err := t.conn.Channel()
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

	t.log.Debug().
		Str("queue", cfg.QueueName).
		Str("exchange", cfg.ExchangeName).
		Str("routing_key", cfg.RoutingKey).
		Msg("Queue bound to exchange")

	return nil
}

// SetupTopology declares exchange, queue and binds them together.
func (t *Topology) SetupTopology(
	exchange ExchangeConfig,
	queue QueueConfig,
	binding BindingConfig,
) error {
	if err := t.DeclareExchange(exchange); err != nil {
		return fmt.Errorf("failed to setup topology: %w", err)
	}

	if _, err := t.DeclareQueue(queue); err != nil {
		return fmt.Errorf("failed to setup topology: %w", err)
	}

	if err := t.BindQueue(binding); err != nil {
		return fmt.Errorf("failed to setup topology: %w", err)
	}

	t.log.Info().
		Str("exchange", exchange.Name).
		Str("queue", queue.Name).
		Str("routing_key", binding.RoutingKey).
		Msg("Topology setup complete")

	return nil
}

// SetupDLQ creates DLQ infrastructure for a queue.
// Naming: {queue_name}.dlx for exchange, {queue_name}.dlq for queue.
// DLQ is always classic queue type with fanout exchange.
// Returns queue config with x-dead-letter-exchange configured.
func (t *Topology) SetupDLQ(queueCfg QueueConfig) (QueueConfig, error) {
	dlxName := queueCfg.Name + dlxSuffix
	dlqName := queueCfg.Name + dlqSuffix

	dlxCfg := NewExchangeConfig(dlxName, ExchangeTypeFanout)
	if err := t.DeclareExchange(dlxCfg); err != nil {
		return QueueConfig{}, fmt.Errorf("failed to setup DLQ: %w", err)
	}

	dlqCfg := NewQueueConfig(dlqName, QueueTypeClassic)
	if _, err := t.DeclareQueue(dlqCfg); err != nil {
		return QueueConfig{}, fmt.Errorf("failed to setup DLQ: %w", err)
	}

	bindingCfg := NewBindingConfig(dlqName, dlxName, "")
	if err := t.BindQueue(bindingCfg); err != nil {
		return QueueConfig{}, fmt.Errorf("failed to setup DLQ: %w", err)
	}

	t.log.Info().
		Str("dlx", dlxName).
		Str("dlq", dlqName).
		Str("main_queue", queueCfg.Name).
		Msg("DLQ setup complete")

	return queueCfg.WithDeadLetterExchange(dlxName), nil
}

// SetupTopologyWithDLQ declares exchange, queue with DLQ, and binds them together.
// This is a convenience method that combines SetupDLQ and SetupTopology.
func (t *Topology) SetupTopologyWithDLQ(
	exchange ExchangeConfig,
	queue QueueConfig,
	binding BindingConfig,
) error {
	queueWithDLQ, err := t.SetupDLQ(queue)
	if err != nil {
		return fmt.Errorf("failed to setup topology with DLQ: %w", err)
	}

	if err = t.SetupTopology(exchange, queueWithDLQ, binding); err != nil {
		return fmt.Errorf("failed to setup topology with DLQ: %w", err)
	}

	t.log.Info().
		Str("exchange", exchange.Name).
		Str("queue", queue.Name).
		Str("routing_key", binding.RoutingKey).
		Msg("Topology with DLQ setup complete")

	return nil
}

// DeleteExchange deletes an exchange.
func (t *Topology) DeleteExchange(name string, ifUnused bool) error {
	ch, err := t.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err = ch.ExchangeDelete(name, ifUnused, false); err != nil {
		return fmt.Errorf("%w %s: %w", ErrDeleteExchange, name, err)
	}

	t.log.Debug().
		Str("exchange", name).
		Msg("Exchange deleted")

	return nil
}

// DeleteQueue deletes a queue.
func (t *Topology) DeleteQueue(name string, ifUnused, ifEmpty bool) error {
	ch, err := t.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err = ch.QueueDelete(name, ifUnused, ifEmpty, false); err != nil {
		return fmt.Errorf("%w %s: %w", ErrDeleteQueue, name, err)
	}

	t.log.Debug().
		Str("queue", name).
		Msg("Queue deleted")

	return nil
}

// PurgeQueue removes all messages from a queue.
func (t *Topology) PurgeQueue(name string) (int, error) {
	ch, err := t.conn.Channel()
	if err != nil {
		return 0, err
	}
	defer ch.Close()

	count, err := ch.QueuePurge(name, false)
	if err != nil {
		return 0, fmt.Errorf("%w %s: %w", ErrPurgeQueue, name, err)
	}

	t.log.Debug().
		Str("queue", name).
		Int("count", count).
		Msg("Queue purged")

	return count, nil
}

// ExchangeExists checks if an exchange exists.
func (t *Topology) ExchangeExists(name string) (bool, error) {
	ch, err := t.conn.Channel()
	if err != nil {
		return false, err
	}
	defer ch.Close()

	err = ch.ExchangeDeclarePassive(name, "", false, false, false, false, nil)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// QueueExists checks if a queue exists.
func (t *Topology) QueueExists(name string) (bool, error) {
	ch, err := t.conn.Channel()
	if err != nil {
		return false, err
	}
	defer ch.Close()

	_, err = ch.QueueDeclarePassive(name, false, false, false, false, nil)
	if err != nil {
		return false, nil
	}

	return true, nil
}
