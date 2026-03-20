// Package messagebus provides abstractions for message broker operations.
package messagebus

import (
	"context"
	"encoding/json"
	"time"
)

// Message represents a message in the message bus.
type Message struct {
	// ID is the unique identifier of the message.
	ID string `json:"id"`

	// Type is the type of the message (e.g., "recipes.created").
	Type string `json:"type"`

	// Timestamp is the time when the message was created.
	Timestamp time.Time `json:"timestamp"`

	// Source is the service that produced the message.
	Source string `json:"source"`

	// Payload is the message payload as raw JSON.
	Payload json.RawMessage `json:"payload"`

	// Headers contains optional metadata for the message (e.g., tracing).
	Headers map[string]any `json:"headers,omitempty"`
}

// Handler is a function that processes a message.
type Handler func(ctx context.Context, msg Message) error

// Publisher is the interface for publishing messages.
type Publisher interface {
	// Publish sends a message to the specified exchange with the given routing key.
	Publish(ctx context.Context, exchange, routingKey string, msg Message) error
}

// Consumer is the interface for consuming messages.
type Consumer interface {
	// Consume starts consuming messages from the specified queue.
	// The handler is called for each message received.
	Consume(ctx context.Context, queue string, handler Handler) error
}

// MessageBus is the main interface for message broker operations.
type MessageBus interface {
	Publisher
	Consumer

	// Close closes the message bus connection.
	Close() error
}
