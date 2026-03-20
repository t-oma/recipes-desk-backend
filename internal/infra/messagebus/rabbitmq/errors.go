// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"errors"
)

// Connection errors.
var (
	// ErrConnectionClosed is returned when the connection is closed.
	ErrConnectionClosed = errors.New("connection closed")
	// ErrNotConnected is returned when there is no active connection.
	ErrNotConnected = errors.New("not connected to RabbitMQ")
	// ErrDial is returned when the connection cannot be established.
	ErrDial = errors.New("dial RabbitMQ")
	// ErrPing is returned when the connection cannot be pinged.
	ErrPing = errors.New("ping RabbitMQ")
	// ErrCloseConnection is returned when the connection cannot be closed.
	ErrCloseConnection = errors.New("close RabbitMQ connection")
	// ErrConnect is returned when the connection cannot be established.
	ErrConnect = errors.New("connect to RabbitMQ")
)

// Channel errors.
var (
	// ErrChannelClosed is returned when the channel is closed.
	ErrChannelClosed = errors.New("channel closed")
	// ErrCreateChannel is returned when the channel cannot be created.
	ErrCreateChannel = errors.New("create channel")
)

var (
	// ErrPublishTimeout is returned when publishing times out.
	ErrPublishTimeout = errors.New("publish timeout")
	// ErrNackReceived is returned when the broker negatively acknowledges a message.
	ErrNackReceived = errors.New("nack received from broker")
)

// Topology errors.
var (
	// ErrDeclareExchange is returned when an exchange cannot be declared.
	ErrDeclareExchange = errors.New("declare exchange")
	// ErrDeleteExchange is returned when an exchange cannot be deleted.
	ErrDeleteExchange = errors.New("delete exchange")
	// ErrDeclareQueue is returned when a queue cannot be declared.
	ErrDeclareQueue = errors.New("declare queue")
	// ErrDeleteQueue is returned when a queue cannot be deleted.
	ErrDeleteQueue = errors.New("delete queue")
	// ErrPurgeQueue is returned when a queue cannot be purged.
	ErrPurgeQueue = errors.New("purge queue")
	// ErrBindQueue is returned when a queue cannot be bound to an exchange.
	ErrBindQueue = errors.New("bind queue")
)
