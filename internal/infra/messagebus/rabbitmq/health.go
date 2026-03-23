// Package rabbitmq provides RabbitMQ implementation of the message bus.
package rabbitmq

import (
	"context"
	"fmt"
	"time"
)

// HealthChecker provides health check functionality for RabbitMQ.
type HealthChecker struct {
	conn *Connection
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(conn *Connection) *HealthChecker {
	return &HealthChecker{conn: conn}
}

// Health performs a health check on the RabbitMQ connection.
func (h *HealthChecker) Health(ctx context.Context) error {
	if err := h.conn.Ping(ctx); err != nil {
		return fmt.Errorf("rabbitmq health check failed: %w", err)
	}
	return nil
}

// Status returns the current health status.
func (h *HealthChecker) Status() map[string]interface{} {
	return map[string]any{
		"connected": h.conn.IsConnected(),
		"timestamp": time.Now().UTC(),
	}
}
