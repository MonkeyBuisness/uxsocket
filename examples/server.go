package main

import (
	"github.com/MonkeyBuisness/uxsocket"
	"log"
	"time"
)

func echoServer(s *uxsocket.Server) {
	for {
		s.Write([]byte("hello"))
		time.Sleep(time.Second)
	}
}

func main() {
	var server uxsocket.Server

	// start echo
	go echoServer(&server)

	if err := server.Listen("/tmp/test.sock"); err != nil {
		log.Println(err)
	}
}
