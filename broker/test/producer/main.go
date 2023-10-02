package main

import (
	"log"
	"rabbitmqv2/broker/lib"
	"time"
)

func main() {
	logger := NewLogger(projectName, connName)

	conn, err := lib.New(connName)
	if err != nil {
		log.Println(err)
		return
	}
	conn.Logger(logger)
	closeFunc, err := conn.Connect()
	if err != nil {
		return
	}
	defer closeConnections(closeFunc)
	// create exchange
	ex := lib.NewExchange(exchangeName, lib.Direct, true)
	q1, err := lib.NewQueue(queueNameUpdateQuantity, routingKeyNameUpdateQuantity, ex.Name)
	if err != nil {
		log.Println("error in queue creation: queue.NewQueue()", err)
		return
	}
	q2, err := lib.NewQueue(queueNameSetInWayStatus, routingKeyNameSetInWayStatus, ex.Name)
	if err != nil {
		log.Println("error in queue creation: queue.NewQueue()", err)
		return
	}
	// set exchange and []queses to connection
	conn.SetExchange(ex).SetQueues(q1, q2)
	// declare exchange and than declare and bind queues
	err = conn.DeclareAndBind()
	if err != nil {
		return
	}
	forever := make(chan struct{})

	err = updateQuantity()
	if err != nil {
		log.Println("update_quantity publish:", err)
		return
	}
	time.Sleep(time.Second * 10)
	err = setInWayStatus()
	if err != nil {
		log.Println("setInWayStatus publish:", err)
		return
	}
	<-forever
}

func updateQuantity() error {
	conn, err := lib.Get(connName)
	if err != nil {
		return err
	}
	err = conn.Publish(routingKeyNameUpdateQuantity, []byte("update quantity 3"))
	if err != nil {
		return err
	}
	return nil
}

func setInWayStatus() error {
	conn, err := lib.Get(connName)
	if err != nil {
		return err
	}
	err = conn.Publish(routingKeyNameSetInWayStatus, []byte("set in way status: Transit"))
	if err != nil {
		return err
	}
	return nil
}
