package lib

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func GetConnection() (*amqp.Connection, error) {
	dns := fmt.Sprintf("amqp://%s:%s@%s:%s/", "user", "password", "localhost", "5672")
	config := amqp.Config{Heartbeat: time.Second * 10}
	conn, err := amqp.DialConfig(dns, config)
	//conn, err := amqp.Dial(dns)
	if err != nil {
		return nil, err
	}

	return conn, nil

	//conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	//defer conn.Close()
}
