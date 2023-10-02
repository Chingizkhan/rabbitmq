package rabbit

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type ClientConsumerFunc func(c Connection, q string, deliveries <-chan amqp.Delivery)

type ClientConsumer struct {
	EventKey string
	Queue    string
	Fn       ClientConsumerFunc
}

func combine(consumers ...*ClientConsumer) map[string]ClientConsumerFunc {
	var relations = make(map[string]ClientConsumerFunc)

	for _, c := range consumers {
		relations[c.Queue] = c.Fn
	}

	return relations
}

type PayClientReq struct {
	Sum        float64 `json:"sum"`
	UserId     uint64  `json:"user_id"`
	SupplierId string  `json:"supplier_uuid"`
	Comment    string  `json:"comment"`
}

func newLeopartPayConsumer() *ClientConsumer {
	//funcName := "LeopartPayConsumer()"

	return &ClientConsumer{
		EventKey: EventLeopartPayClient,
		Queue:    QueueLeopartPayClient,
		Fn: func(c Connection, q string, deliveries <-chan amqp.Delivery) {
			for d := range deliveries {
				//ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

				props := &PayClientReq{}
				err := json.Unmarshal(d.Body, props)
				if err != nil {
					//cancel()
					continue
				}

				log.Printf("%#v", props)

				//err = leopart.PayClient(ctx, leopart.PayClientReq{
				//	Sum:        props.Sum,
				//	UserId:     props.UserId,
				//	SupplierId: props.SupplierId,
				//	Comment:    props.Comment,
				//})
				//if err != nil {
				//	golemService.RegisterError(funcName, fmt.Sprintf("leopart.PayClient(). "+string(d.Body)), err)
				//	cancel()
				//	continue
				//}

				d.Ack(false)
				//cancel()
			}
		},
	}
}
