package main

import (
	"log"
	"rabbitmqv2/broker/lib"
)

func closeConnections(fn lib.CloseFunc) {
	err := fn()
	if err != nil {
		log.Println("can not close connections", err)
		return
	}
}

const (
	connName                     = "leopart_http_producer"
	exchangeName                 = "leoaprt_http"
	routingKeyNameUpdateQuantity = "update_quantity"
	routingKeyNameSetInWayStatus = "set_in_way_status"
	queueNameUpdateQuantity      = exchangeName + "." + routingKeyNameUpdateQuantity
	queueNameSetInWayStatus      = exchangeName + "." + routingKeyNameSetInWayStatus
	projectName                  = "rabbitmq@v2"
)
