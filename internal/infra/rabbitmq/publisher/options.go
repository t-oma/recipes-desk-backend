package publisher

import "recipes-desk/internal/infra/rabbitmq"

type OptionFunc func(o *Options)

type Options struct {
	Exchange rabbitmq.ExchangeOptions
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
		Exchange: rabbitmq.ExchangeOptions{
			Name:    "",
			Kind:    "direct",
			Durable: false,
			Declare: false,
		},
	}
}

func ensureOptions(o *Options) {
	if o.Exchange.Kind == "" {
		o.Exchange.Kind = "direct"
	}
}
