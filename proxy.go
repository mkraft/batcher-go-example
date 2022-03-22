package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type publisher interface {
	publish(*event)
}

type handler struct {
	// maxSize is the maximum allowable size for any given queue name.
	maxSize int

	// waitDuration is the maximum duration that a queue will wait for another message to come in to
	// the queue sharing the same name.
	waitDuration time.Duration

	// matchCriteria returns whether the given WebSocket event matches the the intended criteria of
	// the queue instance, and if so the name of the queue.
	matchCriteria func(*event) (string, bool)

	// batchReducer takes the list of enqueued messages and combines them into one or many messages
	// using whatever logic it wants.
	batchReducer reducerFunc
}

type reducerFunc func(events []*event) []*event

type proxy struct {
	in chan *event
}

func (p *proxy) startCoordinator(ctx context.Context, handlers []*handler, out chan []*event, done chan bool) {
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	queues := make(map[string]chan *event)

	for {
		select {
		case e := <-p.in:
			var evtHasHandler bool

			for _, handler := range handlers {
				queueName, matchesHandler := handler.matchCriteria(e)
				if !matchesHandler {
					continue
				}

				evtHasHandler = true

				mu.Lock()
				queueIn, queueExists := queues[queueName]
				mu.Unlock()
				if !queueExists {
					log.Printf("creating new named queue: %s", queueName)

					queueIn = make(chan *event, handler.maxSize)
					queueDone := make(chan bool)
					wg.Add(1)
					go handleQueue(ctx, queueIn, out, queueDone, handler.waitDuration, handler.maxSize, handler.batchReducer)
					mu.Lock()
					queues[queueName] = queueIn
					mu.Unlock()

					// wait for queue done channel
					go func(name string) {
						select {
						case <-queueDone:
							mu.Lock()
							delete(queues, name)
							mu.Unlock()
							wg.Done()
						}
					}(queueName)
				}

				// send the event through to the queue handler (existing or new)
				queueIn <- e
			}

			if !evtHasHandler {
				out <- []*event{e}
			}

		case <-ctx.Done():
			// wait until all of the named queues are done
			wg.Wait()
			done <- true
			log.Print("orchestrator done: context")
			return
		}
	}
}

func newProxy(ctx context.Context, pub publisher, handlers []*handler) (*proxy, chan bool) {
	out := make(chan []*event)
	done := make(chan bool)
	newProxy := &proxy{in: make(chan *event)}
	go newProxy.startCoordinator(ctx, handlers, out, done)
	go listenForOutEvents(out, pub)
	return newProxy, done
}

func listenForOutEvents(out chan []*event, pub publisher) {
	for e := range out {
		for _, evt := range e {
			pub.publish(evt)
		}
	}
}

func handleQueue(ctx context.Context, in chan *event, out chan []*event, done chan bool, waitDur time.Duration, maxBatchCount int, reduce reducerFunc) {
	queue := []*event{}
	timeout := time.After(waitDur)

	queueDone := func(queueLen int, reason string) {
		switch queueLen {
		case 0:
			// don't send anything on the out channel
		case 1:
			out <- queue
		default:
			out <- reduce(queue)
		}
		done <- true
		log.Printf("queue done: %s", reason)
	}

	for {
		select {
		case evt := <-in:
			queue = append(queue, evt)
			if len(queue) >= maxBatchCount {
				out <- reduce(queue)
				done <- true
				log.Print("queue done: max size")
				return
			}
		case <-ctx.Done():
			queueDone(len(queue), "context")
			return
		case <-timeout:
			queueDone(len(queue), "timeout")
			return
		}
	}
}

func (p *proxy) publish(evt *event) {
	p.in <- evt
}
