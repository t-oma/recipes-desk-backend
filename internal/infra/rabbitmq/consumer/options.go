package consumer

import (
	"context"
	"time"

	"recipes-desk/internal/infra/rabbitmq"
)

type OptionFunc func(o *Options)

type Options struct {
	Queue      string
	Handler    HandlerFunc
	RoutingKey string
	MaxRetries int
	RetryTTLs  []time.Duration // e.g., 5s, 30s, 2m

	Exchange rabbitmq.ExchangeOptions
}

func (o Options) GetRetryTTL(retry int) time.Duration {
	if retry-1 >= len(o.RetryTTLs) {
		return o.RetryTTLs[len(o.RetryTTLs)-1]
	}
	return o.RetryTTLs[retry-1]
}

func WithHandler(handler HandlerFunc) OptionFunc {
	return func(o *Options) {
		o.Handler = handler
	}
}

func WithRoutingKey(routingKey string) OptionFunc {
	return func(o *Options) {
		o.RoutingKey = routingKey
	}
}

func WithMaxRetries(maxRetries int) OptionFunc {
	return func(o *Options) {
		o.MaxRetries = maxRetries
	}
}

func WithRetryTTLs(retryTTLs []time.Duration) OptionFunc {
	return func(o *Options) {
		o.RetryTTLs = retryTTLs
	}
}

func WithExchangeName(name string) OptionFunc {
	return func(o *Options) {
		o.Exchange.Name = name
	}
}

func WithExchangeKind(kind string) OptionFunc {
	return func(o *Options) {
		o.Exchange.Kind = kind
	}
}

func WithExchangeDeclare(o *Options) {
	o.Exchange.Declare = true
}

func WithExchangeDurable(o *Options) {
	o.Exchange.Durable = true
}

func getDefaultOptions() Options {
	return Options{
		Queue:      "",
		Handler:    nil,
		RoutingKey: "",
		MaxRetries: 3,
		RetryTTLs:  []time.Duration{5 * time.Second, 30 * time.Second, 90 * time.Second},

		Exchange: rabbitmq.ExchangeOptions{
			Name:    "",
			Kind:    "direct",
			Durable: false,
			Declare: false,
		},
	}
}

func ensureOptions(o *Options, queue string) {
	if o.Queue == "" {
		o.Queue = queue
	}
	if o.Handler == nil {
		o.Handler = func(ctx context.Context, routingKey string, body []byte) error {
			return nil
		}
	}
	if o.MaxRetries <= 0 {
		o.MaxRetries = getDefaultOptions().MaxRetries
	}
	if len(o.RetryTTLs) == 0 {
		o.RetryTTLs = getDefaultOptions().RetryTTLs
	}
}
