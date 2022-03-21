package main

import (
	"log"
	"math/rand"
)

type eventType string

const (
	typing            eventType = "typing"
	teamJoin          eventType = "team_join"
	multipleTeamJoins eventType = "team_join_multiple"
)

// event represents a websocket event
type event struct {
	name   eventType
	teamID string
	data   interface{}
}

// app simulates an application
type app struct{}

// publish simulates an app method that is invoked by backend clients to publish websocket events
func (a *app) publish(evt *event) {
	log.Printf("event sent to clients: %s, %+v\n", evt.name, evt)
}

// batchTeamJoins is some example business logic about how to combine multiple
// teamJoin events into a single multipleTeamJoins event.
func batchTeamJoins(events []*event) *event {
	out := &event{
		name: multipleTeamJoins,
	}
	var userIDs []string
	for _, event := range events {
		if out.teamID == "" {
			out.teamID = event.teamID
		}
		if event.teamID != out.teamID {
			continue
		}
		userID := event.data.(string)
		userIDs = append(userIDs, userID)
	}
	out.data = userIDs
	return out
}

// randomeString is used for generating fake user ids in the example app
func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
