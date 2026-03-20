// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"errors"
	"fmt"
)

// Connection errors.
var (
	ErrConnection = errors.New("connection")
	// ErrConnectionClosed is returned when the connection is closed.
	ErrConnectionClosed = fmt.Errorf("%w: closed", ErrConnection)
	// ErrNotConnected is returned when there is no active connection.
	ErrNotConnected = fmt.Errorf("%w: not connected", ErrConnection)
)

// Topology errors.
var (
	// ErrTopology is returned when there is an error in the topology.
	ErrTopology = errors.New("topology")
	// ErrDeclareExchange is returned when an exchange cannot be declared.
	ErrDeclareExchange = fmt.Errorf("%w: declare exchange", ErrTopology)
	// ErrDeleteExchange is returned when an exchange cannot be deleted.
	ErrDeleteExchange = fmt.Errorf("%w: delete exchange", ErrTopology)
	// ErrDeclareQueue is returned when a queue cannot be declared.
	ErrDeclareQueue = fmt.Errorf("%w: declare queue", ErrTopology)
	// ErrDeleteQueue is returned when a queue cannot be deleted.
	ErrDeleteQueue = fmt.Errorf("%w: delete queue", ErrTopology)
	// ErrPurgeQueue is returned when a queue cannot be purged.
	ErrPurgeQueue = fmt.Errorf("%w: purge queue", ErrTopology)
	// ErrBindQueue is returned when a queue cannot be bound to an exchange.
	ErrBindQueue = fmt.Errorf("%w: bind queue", ErrTopology)
)

var (
	// ErrPublishTimeout is returned when publishing times out.
	ErrPublishTimeout = errors.New("publish timeout")
	// ErrNackReceived is returned when the broker negatively acknowledges a message.
	ErrNackReceived = errors.New("nack received from broker")
)
