package main

import (
	"log"
	"rabbitmqv2/broker/lib"
)

func main() {
	logger := NewLogger(projectName, connName)

	// create or get existing connection
	conn, err := lib.New(connName)
	if err != nil {
		log.Println(err)
		return
	}
	conn.Logger(logger)
	closeConn, err := conn.Connect()
	if err != nil {
		return
	}
	// close channel and connection
	defer closeConnections(closeConn)
	if err != nil {
		return
	}
	// create exchange
	ex := lib.NewExchange(exchangeName, lib.Direct, true)
	// create queue
	q1, err := lib.NewQueue(queueNameUpdateQuantity, routingKeyNameUpdateQuantity, ex.Name)
	if err != nil {
		return
	}
	q1.WithCallback(updateQuantity)

	q2, err := lib.NewQueue(queueNameSetInWayStatus, routingKeyNameSetInWayStatus, ex.Name)
	if err != nil {
		log.Println("error in queue creation: queue.NewQueue()", err)
		return
	}
	q2.WithCallback(setInWayStatus)

	// set exchange and []queses to connection
	conn.SetExchange(ex).SetQueues(q1, q2)
	// declare exchange and than declare and bind queues
	err = conn.DeclareAndBind()
	if err != nil {
		return
	}
	// endless process to consume all messages
	err = conn.Consume()
	if err != nil {
		return
	}
}
