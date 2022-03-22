package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testApp struct {
	events []*event
	mu     sync.Mutex
}

func (t *testApp) publish(evt *event) {
	t.addEvent(evt)
}

func (t *testApp) addEvent(evt *event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, evt)
}

func (t *testApp) getEvents() []*event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.events
}

func TestPublish(t *testing.T) {
	t.Run("an event without a handler bypasses the reducer", func(t *testing.T) {
		fakeApp := &testApp{}
		fakeEvtName := eventType("fakeevent")
		ctx, cancel := context.WithCancel(context.Background())
		prox, done := newProxy(ctx, fakeApp, []*handler{})
		eventWithoutHandler := &event{name: fakeEvtName}
		prox.publish(eventWithoutHandler)
		cancel()
		for {
			select {
			case <-done:
				require.Equal(t, fakeApp.getEvents()[0].name, fakeEvtName)
				return
			}
		}
	})

	t.Run("single event get published as-is after timeout", func(t *testing.T) {
		fakeApp := &testApp{}
		fakeEvtName := eventType("fakeevent")
		testHandler := &handler{
			maxSize:      100,
			waitDuration: time.Millisecond,
			matchCriteria: func(evt *event) (string, bool) {
				if evt.name != fakeEvtName {
				}
				return "testQueue", true
			},
			batchReducer: func(events []*event) []*event {
				return []*event{{name: "shouldnotget"}}
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		prox, done := newProxy(ctx, fakeApp, []*handler{testHandler})
		eventWithHandler := &event{name: fakeEvtName}
		prox.publish(eventWithHandler)
		time.Sleep(10 * time.Millisecond) // to be after the timeout
		cancel()
		for {
			select {
			case <-done:
				require.Equal(t, fakeApp.getEvents()[0].name, fakeEvtName)
				return
			}
		}
	})

	t.Run("multiple events go through the reducer after timeout", func(t *testing.T) {
		fakeApp := &testApp{}
		fakeEvtName := eventType("shouldnotget")
		fakeEvtMultipleName := eventType("fakeeventmultiple")
		testHandler := &handler{
			maxSize:      100,
			waitDuration: time.Millisecond,
			matchCriteria: func(evt *event) (string, bool) {
				if evt.name != fakeEvtName {
				}
				return "testQueue", true
			},
			batchReducer: func(events []*event) []*event {
				return []*event{{name: fakeEvtMultipleName}}
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		prox, done := newProxy(ctx, fakeApp, []*handler{testHandler})
		eventWithHandler := &event{name: fakeEvtName}
		prox.publish(eventWithHandler)
		prox.publish(eventWithHandler)
		time.Sleep(10 * time.Millisecond) // to be after the timeout
		cancel()
		for {
			select {
			case <-done:
				require.Equal(t, fakeApp.getEvents()[0].name, fakeEvtMultipleName)
				return
			}
		}
	})

	t.Run("multiple events go through the reducer after max batch size", func(t *testing.T) {
		fakeApp := &testApp{}
		fakeEvtName := eventType("shouldnotget")
		fakeEvtMultipleName := eventType("fakeeventmultiple")
		testHandler := &handler{
			maxSize:      1,
			waitDuration: time.Millisecond,
			matchCriteria: func(evt *event) (string, bool) {
				if evt.name != fakeEvtName {
				}
				return "testQueue", true
			},
			batchReducer: func(events []*event) []*event {
				return []*event{{name: fakeEvtMultipleName}}
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		prox, done := newProxy(ctx, fakeApp, []*handler{testHandler})
		eventWithHandler := &event{name: fakeEvtName}
		prox.publish(eventWithHandler)
		prox.publish(eventWithHandler)
		cancel()
		for {
			select {
			case <-done:
				require.Equal(t, fakeApp.getEvents()[0].name, fakeEvtMultipleName)
				return
			}
		}
	})
}
