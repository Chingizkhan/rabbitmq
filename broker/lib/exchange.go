package lib

import amqp "github.com/rabbitmq/amqp091-go"

const (
	Direct  = "direct"
	FanOut  = "fanout"
	Topic   = "topic"
	Headers = "headers"
)

type Exchange struct {
	Name, Kind string
	Durable    bool
}

func NewExchange(name, kind string, durable bool) *Exchange {
	return &Exchange{
		Name:    name,
		Kind:    kind,
		Durable: durable,
	}
}

func (e *Exchange) Declare(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(e.Name+postfixDLX, e.Kind, e.Durable, false, false, false, nil)
	if err != nil {
		return err
	}

	return ch.ExchangeDeclare(e.Name, e.Kind, e.Durable, false, false, false, nil)
}
