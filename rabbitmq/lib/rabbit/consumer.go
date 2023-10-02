package rabbit

func Consume() {
	funcName := "Consume()"

	forever := make(chan bool)

	leopartPayConsumer := newLeopartPayConsumer()

	// map[queueName]callbackFunc
	relations := combine(leopartPayConsumer)

	conn := NewConnection(
		"leopart_pay_consumer",
		"transactions",
		[]string{
			leopartPayConsumer.Queue,
		},
		map[string]string{
			leopartPayConsumer.Queue: leopartPayConsumer.EventKey,
		},
	)

	err := conn.Connect()
	failOnError(err, funcName+" failed to connect")

	err = conn.BindQueue()
	failOnError(err, funcName+" failed bindQueue")

	msgs, err := conn.Consume()
	failOnError(err, funcName+" failed consume")

	for q, m := range msgs {
		go conn.HandleConsumedDeliveries(q, m, relations[q])
	}

	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		panic(msg + " : " + err.Error())
	}
}
