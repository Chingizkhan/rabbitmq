package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

// rabbitmqctl list_bindings
// rabbitmqctl list_queues name messages_ready messages_unacknowledged
// rabbitmqctl list_exchanges

func GetConnection() (*amqp.Connection, error) {
	//params := config.Get().Rabbit
	//dns := fmt.Sprintf("amqp://%s:%s@%s:%s/", params.User, params.Password, params.Host, params.Port)
	dns := fmt.Sprintf("amqp://%s:%s@%s:%s/", "user", "password", "localhost", "5672")
	conn, err := amqp.Dial(dns)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
