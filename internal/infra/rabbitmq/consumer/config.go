package consumer

import (
	"context"
	"time"
)

// HandlerFunc is the function signature for event handlers.
type HandlerFunc func(ctx context.Context, routingKey string, body []byte) error

type Config struct {
	Exchange   string
	Queue      string
	RoutingKey string
	Handler    HandlerFunc
	MaxRetries int
	RetryTTLs  []time.Duration // e.g., 5s, 30s, 2m
}

func (c Config) WithDefaults() Config {
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if len(c.RetryTTLs) == 0 {
		c.RetryTTLs = []time.Duration{
			5 * time.Second,
			30 * time.Second,
			90 * time.Second,
		}
	}

	return c
}

// GetRetryTTL returns the TTL for the given retry attempt.
func (c Config) GetRetryTTL(retry int) time.Duration {
	if retry-1 >= len(c.RetryTTLs) {
		return c.RetryTTLs[len(c.RetryTTLs)-1]
	}
	return c.RetryTTLs[retry-1]
}
