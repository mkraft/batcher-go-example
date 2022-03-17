package main

import (
	"fmt"
	"time"
)

type eventType string

const (
	typing   eventType = "typing"
	teamJoin eventType = "team_join"
)

type event struct {
	name     eventType
	messages []string
}

type app struct{}

func (a *app) publish(evt *event) {
	fmt.Printf("%s: %v\n", evt.name, len(evt.messages))
}

type publisher interface {
	publish(*event)
}

type proxy struct {
	publisher publisher
	in        chan *event
	out       chan *event
	queue     map[eventType][]*event
}

func (p *proxy) publish(evt *event) {
	p.in <- evt
}

func newProxy(pub publisher) *proxy {
	newProxy := &proxy{
		publisher: pub,
		in:        make(chan *event),
		out:       make(chan *event),
		queue:     make(map[eventType][]*event),
	}
	go newProxy.listenAndServe()
	return newProxy
}

// evenReducer is a function that's invoked on all batches of events (even batches with just a single item) prior to being published
// by the publisher. It's an opportunity for clients to combine, transform, or decorate events via the proxy however they'd like.
type eventReducer func(list []*event) *event

// eventHandlerFunc is a helper to create a handler given a common case, where the caller just wants to configure the given parameters
func eventHandlerFunc(waitDur time.Duration, name eventType, maxBatchCount int, reduce eventReducer) eventHandler {
	return func(in chan *event, out chan *event) {
		queue := []*event{}
		timeout := time.After(waitDur)
		for {
			select {
			case evt := <-in:
				if evt.name != name {
					continue
				}
				queue = append(queue, evt)
				if len(queue) >= maxBatchCount {
					out <- reduce(queue)
					queue = []*event{}
				} else {
					queue = append(queue, evt)
				}
			case <-timeout:
				timeout = time.After(waitDur)
				if len(queue) > 0 {
					out <- reduce(queue)
					queue = []*event{}
				}
			}
		}
	}
}

type eventHandler func(in chan *event, out chan *event)

var handlers map[eventType]eventHandler

func init() {
	// eventReducer represents some business logic about how to batch multiple events into one
	eventReducer := func(events []*event) *event {
		if len(events) < 1 {
			return nil
		}
		combined := &event{name: events[0].name}
		for _, evt := range events {
			combined.messages = append(combined.messages, evt.messages...)
		}
		return combined
	}

	handlers = map[eventType]eventHandler{
		typing:   eventHandlerFunc(2*time.Second, typing, 100, eventReducer),
		teamJoin: eventHandlerFunc(4*time.Second, teamJoin, 20, eventReducer),
	}
}

func (p *proxy) listenAndServe() error {
	for _, evt := range []eventType{typing, teamJoin} {
		if handler, has := handlers[evt]; has {
			go handler(p.in, p.out) // each event type gets its own goroutine
		}
	}

	// as the handlers send events back out then invoke the publisher's publish method, in the case (*app).publish
	for e := range p.out {
		p.publisher.publish(e)
	}

	return nil
}

func main() {
	myApp := &app{}

	// note below that (*proxy).publish and (*app).publish have the same signature so the substitution is seemless to the client
	myProxy := newProxy(myApp)

	// simulate client 1 triggering events
	go func() {
		for range time.Tick(50 * time.Millisecond) {
			myProxy.publish(&event{
				name:     typing,
				messages: []string{"t"},
			})
		}
	}()

	// simulate client 2 triggering events
	for range time.Tick(500 * time.Millisecond) {
		myProxy.publish(&event{
			name:     teamJoin,
			messages: []string{"j"},
		})
	}
}
