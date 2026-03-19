package rabbitmq

// Event is the interface for all domain events.
type Event interface {
	Type() string
	RoutingKey() string
}
