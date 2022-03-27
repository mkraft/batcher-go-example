package main

import (
	"context"
	"fmt"
	"time"

	batchelor "github.com/mkraft/batchelorg"
)

func main() {
	joinTeamHandler := &batchelor.Handler{
		Wait: time.Second,
		Match: func(evt *batchelor.Message) (string, bool) {
			if evt.Type != "myMessage" {
				return "", false
			}
			return "myMessages", true
		},
		Reduce: func(messages []*batchelor.Message) *batchelor.Message {
			var combinedData string
			for _, message := range messages {
				combinedData = fmt.Sprintf("%v:%v", combinedData, message.Data)
			}
			return &batchelor.Message{Type: "myCombinedMessages", Data: combinedData}
		},
	}
	handlers := []*batchelor.Handler{joinTeamHandler}
	ctx, cancel := context.WithCancel(context.Background())
	proxy := batchelor.NewProxy(ctx, handlers)

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
			proxy.In(&batchelor.Message{Type: "myMessage", Data: fmt.Sprintf("data%d", i)})
		}
	}()

	<-done
}
