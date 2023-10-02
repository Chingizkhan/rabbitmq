package main

import (
	"errors"
	"log"
	"time"
)

var i = 0

func updateQuantity(body []byte) error {
	log.Println("Salam from updateQuantity func. Body:", string(body))

	if i < 1 {
		i++
		return errors.New("error fuck")
	}

	log.Println("Wait 2 sec")
	time.Sleep(2 * time.Second)
	log.Println("finish")
	log.Println()

	return nil
}

func setInWayStatus(body []byte) error {
	log.Println("Salam from setInWayStatus func. Body:", string(body))

	log.Println("Wait 5 sec")
	time.Sleep(5 * time.Second)
	log.Println("finish")
	log.Println()

	return nil
}
