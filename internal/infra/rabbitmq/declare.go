package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

func DeclareExchange(ch *amqp.Channel, exchange ExchangeOptions) error {
	if !exchange.Declare {
		return nil
	}

	if err := ch.ExchangeDeclare(
		exchange.Name,
		exchange.Kind,
		exchange.Durable,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	return nil
}
