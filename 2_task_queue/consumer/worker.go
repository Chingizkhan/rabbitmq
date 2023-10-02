package main

import (
	"bytes"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

// sudo rabbitmqctl list_queues name messages_ready messages_unacknowledged - check unacknowledged messages

// Round-robin dispatching
// One of the advantages of using a Task Queue is the ability to easily parallelise work.
// If we are building up a backlog of work, we can just add more workers and that way, scale easily.

// By default, RabbitMQ will send each message to the next consumer, in sequence. On average every consumer will
// get the same number of messages. This way of distributing messages is called round-robin.
func main() {
	// Note that we declare the queue here, as well. Because we might start the consumer before the publisher,
	// we want to make sure the queue exists before we try to consume messages from it.
	conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Doing a task can take a few seconds, you may wonder what happens if a consumer starts a long task and it terminates
	// before it completes. With our current code, once RabbitMQ delivers a message to the consumer, it immediately marks
	// it for deletion. In this case, if you terminate a worker, the message it was just processing is lost.
	// The messages that were dispatched to this particular worker but were not yet handled are also lost.
	//
	// But we don't want to lose any tasks. If a worker dies, we'd like the task to be delivered to another worker.
	//
	// In order to make sure a message is never lost, RabbitMQ supports message acknowledgments.
	// An ack(nowledgement) is sent back by the consumer to tell RabbitMQ that a particular message has been received,
	// processed and that RabbitMQ is free to delete it.
	//
	// If a consumer dies (its channel is closed, connection is closed, or TCP connection is lost) without sending an ack,
	// RabbitMQ will understand that a message wasn't processed fully and will re-queue it. If there are other consumers
	// online at the same time, it will then quickly redeliver it to another consumer.
	// That way you can be sure that no message is lost, even if the workers occasionally die.
	//
	// A timeout (30 minutes by default) is enforced on consumer delivery acknowledgement.
	// This helps detect buggy (stuck) consumers that never acknowledge deliveries.
	// You can increase this timeout as described in Delivery Acknowledgement Timeout.
	//
	// In this tutorial we will use manual message acknowledgements by passing a false for the "auto-ack"
	// argument and then send a proper acknowledgment from the worker with d.Ack(false) (this acknowledges a single delivery),
	// once we're done with a task.
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dotCount := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dotCount)
			ti := t * time.Second
			time.Sleep(ti)
			log.Printf("Done in %s", ti)
			d.Ack(false)
		}
	}()
	// Using this code, you can ensure that even if you terminate a worker using CTRL+C while it was processing a message,
	// nothing is lost. Soon after the worker terminates, all unacknowledged messages are redelivered.
	//
	// Acknowledgement must be sent on the same channel that received the delivery.
	// Attempts to acknowledge using a different channel will result in a channel-level protocol exception.
	// See the doc guide on confirmations to learn more.

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
