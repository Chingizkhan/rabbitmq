package lib

import (
	"context"
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	errConnClosed = errors.New("connection closed")
	errChanClosed = errors.New("channel closed")
)

func errConnNotExist(name string) error {
	return errors.New("connection " + name + " does not exist")
}

const headerRetryCount = "retry-count"

var once sync.Once

type Connection struct {
	name      string
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	exchange  *Exchange
	queues    []*Queue
	logger    Logger
	errors    chan error
	broadcast chan struct{}
}

var connectionPool = make(map[string]*Connection)

func New(name string) (*Connection, error) {
	if conn, ok := connectionPool[name]; ok {
		if !conn.Conn.IsClosed() && !conn.Channel.IsClosed() {
			return conn, nil
		}
	}

	if err := validate(name); err != nil {
		return nil, err
	}

	conn := &Connection{
		name:      name,
		Conn:      nil,
		Channel:   nil,
		exchange:  nil,
		queues:    nil,
		errors:    make(chan error, 1),
		broadcast: make(chan struct{}, 1),
	}

	connectionPool[name] = conn

	return conn, nil
}

func validate(name string) error {
	if !strings.Contains(name, "_producer") && !strings.Contains(name, "_consumer") {
		return errors.New("connection must have '_producer' or '_consumer' in name '" + name + "'")
	}
	return nil
}

func Get(name string) (*Connection, error) {
	if conn, ok := connectionPool[name]; ok {
		if !conn.Conn.IsClosed() && !conn.Channel.IsClosed() {
			return conn, nil
		}
	}
	return nil, errConnNotExist(name)
}

type CloseFunc func() error

func (c *Connection) Connect() (CloseFunc, error) {
	const funcName = "Connect()"

	once.Do(func() {
		go c.processErrors()
		// notify broadcast to invoke c.Consume()
		c.broadcast <- struct{}{}
	})

	conn, err := GetConnection()
	if err != nil {
		c.err(funcName, "lib.GetConnection()", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		c.err(funcName, "conn.Channel()", err)
		return nil, err
	}

	go c.listenClosing()

	c.Conn = conn
	c.Channel = ch

	return c.Close, nil
}

func (c *Connection) listenClosing() {
	const funcName = "listenClosing()"

	for {
		select {
		case err := <-c.Conn.NotifyClose(make(chan *amqp.Error)):
			c.err(funcName, errConnClosed.Error(), err)
			c.errors <- errConnClosed
			return
		case err := <-c.Channel.NotifyClose(make(chan *amqp.Error)):
			c.err(funcName, errChanClosed.Error(), err)
			c.errors <- errChanClosed
			return
		}
	}
}

func (c *Connection) processErrors() {
	const funcName = "processErrors()"

	for {
		select {
		case err := <-c.errors:
			c.wait(err.Error())
			err = c.Reconnect()
			if err != nil {
				c.errors <- err
			}

			switch {
			case strings.Contains(c.name, "producer"):
				c.info(funcName, "processErrors -> Producer")
			case strings.Contains(c.name, "consumer"):
				c.info(funcName, "processErrors -> Producer")
				c.broadcast <- struct{}{}
			}
		}
	}
}

func (c *Connection) wait(info string) {
	const funcName = "wait()"

	duration := 10

	c.info(funcName, info+". Waiting for recovery every "+strconv.Itoa(duration)+" second.")
	time.Sleep(time.Duration(duration) * time.Second)
}

func (c *Connection) Reconnect() error {
	const funcName = "Reconnect()"

	c.info(funcName, "start reconnection.")

	err := c.Close()
	if err != nil {
		c.err(funcName, "c.Close()", err)
		return err
	}

	_, err = c.Connect()
	if err != nil {
		c.err(funcName, "c.Connect()", err)
		return err
	}
	err = c.DeclareAndBind()
	if err != nil {
		c.err(funcName, "c.DeclareAndBind()", err)
		return err
	}

	c.info(funcName, "finish reconnecting.")

	return nil
}

func (c *Connection) SetChannel(ch *amqp.Channel) *Connection {
	c.Channel = ch
	return c
}

func (c *Connection) SetExchange(exchange *Exchange) *Connection {
	c.exchange = exchange
	return c
}

func (c *Connection) SetQueues(queues ...*Queue) *Connection {
	c.queues = queues
	return c
}

func (c *Connection) DeclareExchange() error {
	const funcName = "DeclareExchange()"

	err := c.exchange.Declare(c.Channel)
	if err != nil {
		c.err(funcName, "c.exchange.Declare()", err)
		return err
	}

	return nil
}

func (c *Connection) DeclareAndBind() error {
	const funcName = "DeclareAndBind()"

	err := c.DeclareExchange()
	if err != nil {
		return err
	}

	for _, q := range c.queues {
		err = q.Declare(c.Channel)
		if err != nil {
			c.err(funcName, "q.Declare()", err)
			return err
		}

		err = q.Bind(c.Channel)
		if err != nil {
			c.err(funcName, "q.Bind()", err)
			return err
		}
	}

	go c.ping()

	return nil
}

func (c *Connection) Publish(routingKey string, message []byte) error {
	const funcName = "Publish()"

	if !isPing(message) {
		c.info(funcName, "trying to publish message: "+string(message))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if c.Channel.IsClosed() {
		c.err(funcName, "", errChanClosed)
		return errChanClosed
	}

	err := c.Channel.PublishWithContext(ctx,
		c.exchange.Name,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         message,
			Headers: amqp.Table{
				headerRetryCount: int64(0),
			},
		},
	)
	if err != nil {
		c.err(funcName, "c.Channel.PublishWithContext()", err)
		return err
	}

	return nil
}

// todo: 1 подключение и несколько каналов для работы
// раздельное обслуживание консьюмера и продюсера – под каждого свой коннект, а документация на RabbitMQ настойчиво
// не рекомендует плодить подключения и вместо этого использовать каналы (каналы в amqp это легковесные соединения поверх
// TCP-подключения) поверх одного подключения;
func (c *Connection) Consume() error {
	const funcName = "Consume()"

	var i = 0
	closing := make(chan struct{})

	for {
		select {
		case <-c.broadcast:
			for _, q := range c.queues {
				if i < len(c.queues) {
					i++
				} else {
					closing <- struct{}{}
				}
				err := c.processMessages(q, closing)
				if err != nil {
					c.err(funcName, "c.processMessages()", err)
					return err
				}
			}
			c.info(funcName, " [*] Waiting for messages. To exit press CTRL+C")
			//log.Println("active goroutines:", runtime.NumGoroutine())
		}
	}
}

func (c *Connection) processMessages(q *Queue, closing chan struct{}) error {
	const funcName = "processMessages()"

	msgs, err := c.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		c.err(funcName, "c.Channel.Consume()", err)
		return err
	}

	go c.messages(msgs, closing, q)

	return nil
}

func (c *Connection) messages(msgs <-chan amqp.Delivery, closing chan struct{}, q *Queue) {
	const funcName = "messages()"

	for {
		select {
		case <-closing:
			return
		case d, opened := <-msgs:
			if !opened {
				c.info(funcName, "goroutine is closed (!opened)")
				return
			}
			if isPing(d.Body) {
				log.Println("pong")
				err := d.Ack(false)
				if err != nil {
					c.err(funcName, "ping d.Ack()", err)
				}
				continue
			}
			err := q.Callback(d.Body)
			if err != nil {
				c.err(funcName, "q.Callback()", err)
				err = d.Nack(false, false)
				if err != nil {
					c.err(funcName, "d.Nack()", err)
				}
			} else {
				err = d.Ack(false)
				if err != nil {
					c.err(funcName, "d.Ack()", err)
				}
			}
		}
	}
}

func (c *Connection) err(funcName, info string, err error) {
	if c.logger != nil {
		c.logger.Error(funcName, info, err)
	}
}

func (c *Connection) info(funcName, info string) {
	if c.logger != nil {
		c.logger.Info(funcName, info)
	}
}

func (c *Connection) Close() error {
	const funcName = "Close()"

	if !c.Channel.IsClosed() {
		err := c.Channel.Close()
		if err != nil {
			c.err(funcName, "c.Channel.Close()", err)
			return err
		}
	}

	if !c.Conn.IsClosed() {
		err := c.Conn.Close()
		if err != nil {
			c.err(funcName, "c.Conn.Close()", err)
			return err
		}
	}

	return nil
}

func (c *Connection) Logger(l Logger) {
	c.logger = l
}
