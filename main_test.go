package main

import (
	"sync"
)

type testApp struct {
	events []*event
	mu     sync.Mutex
}

func (t *testApp) publish(evt *event) {
	t.events = append(t.events, evt)
}

func (t *testApp) reset() {
	t.events = []*event{}
}

// func TestPublish(t *testing.T) {
// 	fakeApp := &testApp{}

// 	// fakeReducer just returns the first event with the message changed for testing
// 	fakeReducer := func(events []*event) *event {
// 		first := events[0]
// 		first.messages[0] = fmt.Sprintf("proxied_%s", first.messages[0])
// 		return first
// 	}

// 	handlers := map[eventType]eventHandler{
// 		typing: eventHandlerFunc(0, typing, 100, fakeReducer),
// 	}
// 	prox := newProxy(context.Background(), fakeApp, handlers)

// 	t.Run("an event without a handler bypasses the reducer", func(t *testing.T) {
// 		defer fakeApp.reset()
// 		eventWithoutHandler := &event{
// 			name:     "test",
// 			messages: []string{"message1"},
// 		}
// 		prox.publish(eventWithoutHandler)
// 		require.Contains(t, fakeApp.events, eventWithoutHandler)
// 		require.Equal(t, "message1", fakeApp.events[0].messages[0])
// 	})

// 	t.Run("an event with a handler is processed via the reducer", func(t *testing.T) {
// 		defer fakeApp.reset()
// 		eventWithHandler := &event{
// 			name:     typing,
// 			messages: []string{"message1"},
// 		}
// 		prox.publish(eventWithHandler)
// 		time.Sleep(time.Millisecond)
// 		require.Contains(t, fakeApp.events, eventWithHandler)
// 		require.Equal(t, "proxied_message1", fakeApp.events[0].messages[0])
// 	})

// 	t.Run("multiple events via single handler are handled", func(t *testing.T) {
// 		defer fakeApp.reset()
// 		evt1 := &event{name: typing, messages: []string{"message1"}}
// 		prox.publish(evt1)
// 		evt2 := &event{name: typing, messages: []string{"message2"}}
// 		prox.publish(evt2)
// 		time.Sleep(100 * time.Millisecond)
// 		require.Contains(t, fakeApp.events, evt1)
// 		require.Contains(t, fakeApp.events, evt2)
// 		require.Equal(t, "proxied_message1", fakeApp.events[0].messages[0])
// 		require.Equal(t, "proxied_message2", fakeApp.events[1].messages[0])
// 	})
// }
