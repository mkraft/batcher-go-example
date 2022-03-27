package main

import (
	"context"
	"fmt"
	"time"

	batchelor "github.com/mkraft/batchelorg"
)

type message struct {
	id   string
	data interface{}
}

func (m *message) Type() string {
	return m.id
}

func (m *message) Data() interface{} {
	return m.data
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
		Reduce: func(messages []batchelor.Message) batchelor.Message {
			var combinedData string
			for _, message := range messages {
				combinedData = fmt.Sprintf("%v:%v", combinedData, message.Data())
			}
			return &message{id: "myCombinedMessages", data: combinedData}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	proxy := batchelor.NewProxy(ctx, []*batchelor.Handler{myMessageHandler})

	done := make(chan bool)

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
		done <- true
	}()

	go func() {
		for msg := range proxy.Out {
			fmt.Println(msg)
		}
	}()

	go func() {
		for i := 0; ; i++ {
			time.Sleep(250 * time.Millisecond)
			proxy.In(&message{id: "myMessage", data: fmt.Sprintf("data%d", i)})
		}
	}()

	<-done
}
