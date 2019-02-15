package main

import (
	"github.com/MonkeyBuisness/uxsocket"
	"log"
	"time"
)

func reader(c *uxsocket.Client) error {
	buf := make([]byte, 512)
	for {
		n, err := c.Read(buf[:])
		if err != nil {
			c.Close()
			return err
		}

		log.Println("Client got: ", string(buf[0:n]))
	}
}

func main() {
	// run listener
	for {
		if client, err := uxsocket.NewClient("/tmp/test.sock"); err != nil {
			log.Println(err)
			// reconnect after second
			time.Sleep(time.Second)
		} else {
			go reader(client)
			client.Listen()
		}
	}
}
