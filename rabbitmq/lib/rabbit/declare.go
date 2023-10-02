package rabbit

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareQueue(ch *amqp.Channel, name string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
}

func DeclareExchange(ch *amqp.Channel, exchange, kind string) error {
	if kind == "" {
		kind = "direct"
	}

	return ch.ExchangeDeclare(
		exchange,
		kind,
		true,
		false,
		false,
		false,
		nil,
	)
}
