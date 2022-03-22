package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	joinTeamHandler := &handler{
		maxSize: 10,

		waitDuration: 3 * time.Second,

		matchCriteria: func(evt *event) (string, bool) {
			if evt.name != teamJoin {
				return "", false
			}
			queueName := fmt.Sprintf("team_join:team/%s", evt.teamID)
			return queueName, true
		},

		batchReducer: func(events []*event) []*event {
			return []*event{batchTeamJoins(events)}
		},
	}

	handlers := []*handler{joinTeamHandler}

	myApp := &app{}

	ctx, cancel := context.WithCancel(context.Background())

	myProxy, done := newProxy(ctx, myApp, handlers)

	go func() {
		time.Sleep(10 * time.Second)
		log.Print("server triggered a cancellation")
		cancel()
		for {
			select {
			case <-done:
				os.Exit(0)
				return
			}
		}
	}()

	// simulate someone joining teamA
	go func() {
		log.Print("server publishing a teamJoin event")
		myProxy.publish(&event{
			name:   teamJoin,
			teamID: "teamA",
			data:   randomString(26),
		})
	}()

	// simulate some event that doesn't use the handlers
	go func() {
		log.Print("server publishing an event without a handler")
		myProxy.publish(&event{name: "foobar"})
	}()

	// simulate multiple people joining teamB
	for range time.Tick(time.Second) {
		log.Print("server publishing a teamJoin event")
		myProxy.publish(&event{
			name:   teamJoin,
			teamID: "teamB",
			data:   randomString(26),
		})
	}
}
