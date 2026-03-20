// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import "errors"

var (
	// ErrConnectionClosed is returned when the connection is closed.
	ErrConnectionClosed = errors.New("connection closed")

	// ErrChannelClosed is returned when the channel is closed.
	ErrChannelClosed = errors.New("channel closed")

	// ErrPublishTimeout is returned when publishing times out.
	ErrPublishTimeout = errors.New("publish timeout")

	// ErrNackReceived is returned when the broker negatively acknowledges a message.
	ErrNackReceived = errors.New("nack received from broker")

	// ErrNotConnected is returned when there is no active connection.
	ErrNotConnected = errors.New("not connected to RabbitMQ")

	// ErrInvalidExchange is returned when exchange configuration is invalid.
	ErrInvalidExchange = errors.New("invalid exchange configuration")

	// ErrInvalidQueue is returned when queue configuration is invalid.
	ErrInvalidQueue = errors.New("invalid queue configuration")
)
