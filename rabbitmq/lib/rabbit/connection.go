package rabbit

import (
	"context"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"rabbitmqv2/rabbitmq"
	"strconv"
	"time"
)

type Connection struct {
	name      string
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	exchange  string
	queues    []string
	err       chan error
	eventKeys map[string]string
}

var (
	connectionPool = make(map[string]*Connection)
)

func NewConnection(name, exchange string, queues []string, eventKeys map[string]string) *Connection {
	if c, ok := connectionPool[name]; ok {
		return c
	}
	c := &Connection{
		exchange:  exchange,
		queues:    queues,
		eventKeys: eventKeys,
		err:       make(chan error),
	}
	connectionPool[name] = c
	return c
}

func GetConnection(name string) *Connection {
	return connectionPool[name]
}

func (c *Connection) Connect() error {
	const funcName = "broker -> Connect()"

	var err error
	c.Conn, err = rabbitmq.GetConnection()
	if err != nil {
		//golemService.RegisterError(funcName, "broker.GetConnection()", err)
		return fmt.Errorf("error in creating rabbitmq connection with : %s", err.Error())
	}
	go func() {
		<-c.Conn.NotifyClose(make(chan *amqp.Error))
		c.err <- errors.New("connection closed")
	}()
	c.Channel, err = c.Conn.Channel()
	if err != nil {
		//golemService.RegisterError(funcName, "c.Conn.Channel()", err)
		return fmt.Errorf("channel: %s", err)
	}
	if err := DeclareExchange(c.Channel, c.exchange, "direct"); err != nil {
		//golemService.RegisterError(funcName, "event.DeclareExchange()", err)
		return fmt.Errorf("error in exchange declare: %s", err)
	}
	return nil
}

// todo: сделать event keys для queue массивом
// todo: убрать queues при создании connection
func (c *Connection) BindQueue() error {
	const funcName = "broker -> BindQueue()"

	for _, q := range c.queues {
		if _, err := DeclareQueue(c.Channel, q); err != nil {
			//golemService.RegisterError(funcName, "event.DeclareExchange()", err)
			return fmt.Errorf("error in declaring the queue %s", err)
		}
		eventKey, ok := c.eventKeys[q]
		if !ok {
			err := fmt.Errorf("event_key for queue [%s] not found", q)
			//golemService.RegisterError(funcName, "c.eventKeys[q]", err)
			return err
		}
		if err := c.Channel.QueueBind(q, eventKey, c.exchange, false, nil); err != nil {
			//golemService.RegisterError(funcName, "c.Channel.QueueBind()", err)
			return fmt.Errorf("queue bind error: %s", err)
		}
	}
	return nil
}

func (c *Connection) Reconnect() error {
	const funcName = "broker -> Reconnect()"

	if err := c.Connect(); err != nil {
		//golemService.RegisterError(funcName, "c.Connect()", err)
		return err
	}
	if err := c.BindQueue(); err != nil {
		//golemService.RegisterError(funcName, "c.BindQueue()", err)
		return err
	}
	return nil
}

func (c *Connection) Publish(ctx context.Context, event string, body []byte) error {
	const funcName = "broker -> Publish()"

	select {
	case err := <-c.err:
		if err != nil {
			err := c.Reconnect()
			if err != nil {
				return err
			}
		}
	default:
	}
	err := c.Channel.PublishWithContext(ctx,
		c.exchange,
		event,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		if errors.Is(err, amqp.ErrClosed) {
			err := c.Reconnect()
			if err != nil {
				return err
			}
		}
		//golemService.RegisterError(funcName, "c.Channel.PublishWithContext()", err)
		return fmt.Errorf("error while publishing: (%s) %s", body, err)
	}
	return nil
}

func (c *Connection) Consume() (map[string]<-chan amqp.Delivery, error) {
	const funcName = "broker -> Consume()"

	m := make(map[string]<-chan amqp.Delivery)

	for _, q := range c.queues {
		msgs, err := c.Channel.Consume(q, "", false, false, false, false, nil)
		if err != nil {
			//golemService.RegisterError(funcName, "c.Channel.PublishWithContext()", err)
			return nil, err
		}
		m[q] = msgs
	}

	return m, nil
}

func wait() {
	duration := 5
	log.Println("Channel is closed. Waiting for recovery every " + strconv.Itoa(duration) + " second")
	time.Sleep(time.Duration(duration) * time.Second)
}

func (c *Connection) HandleConsumedDeliveries(q string, delivery <-chan amqp.Delivery, fn func(Connection, string, <-chan amqp.Delivery)) {
	const funcName = "broker -> HandleConsumedDeliveries()"

	var deliveryProp = delivery

	for {
		go fn(*c, q, deliveryProp)

		if err := <-c.err; err != nil {
			reconnected := make(chan bool, 1)
			msgs := make(map[string]<-chan amqp.Delivery)
			var msgsErr error

		LOOP:
			for {
				select {
				case <-reconnected:
					close(reconnected)
					break LOOP
				default:
					err = c.Reconnect()
					if err != nil {
						//golemService.RegisterError(funcName, "c.Reconnect()", err)
						wait()
						continue LOOP
					}
					msgs, msgsErr = c.Consume()
					if msgsErr != nil {
						//golemService.RegisterError(funcName, "c.Consume()", err)
						wait()
						continue LOOP
					} else {
						reconnected <- true
						continue LOOP
					}
				}
			}
			deliveryProp = msgs[q]
			continue
		}
	}
}
