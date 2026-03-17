package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(config Config) (*amqp.Connection, error) {
	uri := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.VHost,
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return conn, nil
}
