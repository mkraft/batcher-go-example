package main

import (
	"context"
	"fmt"
	"time"

	batchelor "github.com/mkraft/batchelorgo"
)

type message struct {
	id   string
	data interface{}
}

func (m *message) Type() string {
	return m.id
}

func (m *message) String() string {
	return fmt.Sprintf("id: %s, data: %v", m.id, m.data)
}

func main() {
	myMessageHandler := &batchelor.Handler{
		Wait: time.Second,
		Match: func(msg batchelor.Message) (string, bool) {
			if msg.Type() != "myMessage" {
				return "", false
			}
			return "myMessages", true
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	proxy := batchelor.NewProxy(ctx, []*batchelor.Handler{myMessageHandler})

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

	for msg := range proxy.Out {
		fmt.Println(msg)
	}
}
