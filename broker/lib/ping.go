package lib

import (
	"time"
)

func (c *Connection) ping() {
	const funcName = "ping()"

	for {
		time.Sleep(30 * time.Minute)
		err := c.Publish(c.queues[0].RoutingKey, []byte("ping"))
		if err != nil {
			c.err(funcName, "publish()", err)
			return
		}
	}
}

func isPing(msg []byte) bool {
	return string(msg) == "ping"
}
