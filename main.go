package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mkraft/batcher-go"
)

type message struct {
	id   string
	data interface{}
}

func (m *message) String() string {
	return fmt.Sprintf("id: %s, data: %v", m.id, m.data)
}

func main() {
	myMessageHandler := &batcher.Handler{
		Wait: time.Second,
		Match: func(raw interface{}) (string, bool) {
			msg, ok := raw.(*message)
			if !ok {
				log.Print("conversion error")
				return "", false
			}
			if msg.id != "myMessage" {
				return "", false
			}
			return "myMessages", true
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	proxy := batcher.NewBatcher(ctx, []*batcher.Handler{myMessageHandler})

	// set a timeout to shut down the proxy
	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()

	// fake a stream of incoming messages
	go func() {
		for i := 0; ; i++ {
			time.Sleep(250 * time.Millisecond)
			proxy.In(&message{id: "myMessage", data: fmt.Sprintf("data%d", i)})
		}
	}()

	// batches are sent out, as configured by the handler logic
	for batch := range proxy.Out {
		fmt.Println(batch)
	}
}
