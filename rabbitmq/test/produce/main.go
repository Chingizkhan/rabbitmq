package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"rabbitmqv2/rabbitmq/lib/rabbit"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	//err := LeopartPayToClient(ctx, &LeopartPayToClientReq{
	//	Sum:        10000,
	//	UserId:     12,
	//	SupplierId: "hz",
	//	Comment:    "Comment",
	//})
	//if err != nil {
	//	log.Println("err", err)
	//	return
	//}

	err := rabbit.Produce(ctx, rabbit.PublishReq{
		Exchange: "transactions",
		Key:      "EventLeopartPayClient",
		Data: map[string]int{
			"id":    1,
			"level": 1,
		},
	})
	if err != nil {
		log.Println("err on produce", err)
		return
	}
}

type LeopartPayToClientReq struct {
	Sum        float64 `json:"sum"`
	UserId     uint64  `json:"user_id"`
	SupplierId string  `json:"supplier_uuid"`
	Comment    string  `json:"comment"`
}

func LeopartPayToClient(ctx context.Context, req *LeopartPayToClientReq) error {
	conn := rabbit.GetConnection("orderItem_producer")
	if conn == nil {
		//return status.Error(codes.Code(http.StatusInternalServerError),
		//	"can not find connection in 'LeopartPayToClient()'")
		return errors.New("conn is nil")
	}

	js, err := json.Marshal(req)
	if err != nil {
		return err
	}

	err = conn.Publish(ctx, rabbit.EventLeopartPayClient, js)
	if err != nil {
		return err
	}

	return nil
}
