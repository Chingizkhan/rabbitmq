package rabbit

import (
	"context"
	"encoding/json"
	"log"
)

type PublishReq struct {
	ConnName  string
	Exchange  string
	EventKeys map[string]string
	Key       string
	Queues    []string
	Data      interface{}
}

func Produce(ctx context.Context, req PublishReq) error {
	conn := NewConnection(
		req.ConnName,
		req.Exchange,
		req.Queues,
		req.EventKeys,
	)
	if err := conn.Connect(); err != nil {
		return err
	}
	defer conn.Conn.Close()
	defer conn.Channel.Close()

	if err := conn.BindQueue(); err != nil {
		return err
	}

	js, err := json.Marshal(req.Data)
	if err != nil {
		return err
	}

	err = conn.Publish(ctx, req.Key, js)
	if err != nil {
		log.Println("error on publish:", err)
		return err
	}
	log.Println("send ", string(js))

	return nil
}
