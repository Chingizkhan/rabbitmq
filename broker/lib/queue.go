package lib

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

var errEmptyName = errors.New("'name' of the queue is empty")
var errEmptyRoutingKey = errors.New("'routing key' of the queue is empty")
var errEmptyExchangeName = errors.New("'exchange name' of the queue is empty")

type Callback func([]byte) error

type Queue struct {
	Name, RoutingKey, Exchange string
	AmqpQueue                  *amqp.Queue
	Callback                   Callback
}

func NewQueue(name, key, exchange string) (*Queue, error) {
	if err := validateQueue(name, key, exchange); err != nil {
		return nil, err
	}

	return &Queue{
		Name:       name,
		RoutingKey: key,
		Exchange:   exchange,
		Callback:   nil,
	}, nil
}

func validateQueue(name, key, exchange string) error {
	if name == "" {
		return errEmptyName
	}
	if key == "" {
		return errEmptyRoutingKey
	}
	if exchange == "" {
		return errEmptyExchangeName
	}
	return nil
}

func (q *Queue) WithCallback(fn Callback) *Queue {
	q.Callback = fn
	return q
}

func (q *Queue) Declare(ch *amqp.Channel) error {
	argsDLX, args := q.Args()

	queue, err := ch.QueueDeclare(q.Name+postfixDLX, true, false, false, false, argsDLX)
	if err != nil {
		return err
	}

	queue, err = ch.QueueDeclare(q.Name, true, false, false, false, args)
	if err != nil {
		return err
	}

	q.AmqpQueue = &queue
	return nil
}

func (q *Queue) Args() (argsDlx amqp.Table, args amqp.Table) {
	argsDlx = amqp.Table{
		dlxName: q.Exchange,
		dlxTTL:  int64(10000),
	}
	args = amqp.Table{dlxName: q.Exchange + postfixDLX}
	return
}

func (q *Queue) Bind(ch *amqp.Channel) error {
	err := ch.QueueBind(
		q.Name+postfixDLX,
		q.RoutingKey,
		q.Exchange+postfixDLX,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return ch.QueueBind(
		q.Name,       // queue name
		q.RoutingKey, // routing key
		q.Exchange,   // exchange
		false,
		nil,
	)
}
