// Package sse implements a minimal Server-Sent Events broker for pushing
// monitor status changes to connected browser clients in real time.
//
// Internally, updates are published per monitor ID — this keeps the
// primitive granular and ready for a future single-monitor detail page.
// A client can subscribe to one monitor or to a set of monitor IDs (e.g.
// "all monitors belonging to this user") through a single connection;
// the broker fans out to whichever monitor IDs that subscriber cares
// about.
package sse

import (
	"encoding/json"
	"sync"
)

// Event is a single message pushed to subscribers. Name becomes the SSE
// "event:" field (e.g. "status_change"); Data is marshaled to JSON and
// sent as the "data:" field.
type Event struct {
	MonitorID int
	Name      string
	Data      any
}

// subscriber is one open connection's mailbox. ch receives events for any
// monitor ID in monitorIDs. A nil/empty monitorIDs set is never matched —
// subscribers must explicitly list what they want.
type subscriber struct {
	ch         chan Event
	monitorIDs map[int]struct{}
}

// Broker fans out published events to all subscribers interested in the
// relevant monitor ID. Safe for concurrent use.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[*subscriber]struct{}
}

// NewBroker creates an empty broker. One Broker instance should be shared
// across the whole process (constructed once in main/server setup and
// passed to both the worker and the SSE HTTP handler).
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[*subscriber]struct{}),
	}
}

// Subscribe registers a new listener interested in the given monitor IDs
// and returns a channel of events plus an unsubscribe function. Callers
// MUST call unsubscribe (typically via defer) when the connection closes,
// or the subscriber and its channel will leak for the life of the process.
//
// The returned channel is buffered to tolerate a slow consumer for a short
// burst; if the buffer fills, Publish drops the oldest-pending event for
// that subscriber rather than blocking the publisher (a worker goroutine
// must never stall because a browser tab is slow to read).
func (b *Broker) Subscribe(monitorIDs []int) (events <-chan Event, unsubscribe func()) {
	idSet := make(map[int]struct{}, len(monitorIDs))
	for _, id := range monitorIDs {
		idSet[id] = struct{}{}
	}

	sub := &subscriber{
		ch:         make(chan Event, 16),
		monitorIDs: idSet,
	}

	b.mu.Lock()
	b.subscribers[sub] = struct{}{}
	b.mu.Unlock()

	unsubscribe = func() {
		b.mu.Lock()
		delete(b.subscribers, sub)
		b.mu.Unlock()
		close(sub.ch)
	}

	return sub.ch, unsubscribe
}

// Publish sends an event to every current subscriber interested in
// event.MonitorID. Called by the worker after each health check.
func (b *Broker) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for sub := range b.subscribers {
		if _, ok := sub.monitorIDs[event.MonitorID]; !ok {
			continue
		}
		select {
		case sub.ch <- event:
		default:
			// Subscriber's buffer is full (slow consumer). Drop this event
			// for them rather than blocking the publisher — the next
			// dashboard poll-equivalent (reconnect, or the next event) will
			// still bring them back to a correct state since each event
			// carries the monitor's full current status, not a diff.
		}
	}
}

// MarshalData is a convenience for handlers writing the SSE wire format:
// "event: <name>\ndata: <json>\n\n". Returns the JSON-encoded data only;
// callers are responsible for the surrounding SSE framing since that's
// inherently tied to http.ResponseWriter and shouldn't live in this
// transport-agnostic package.
func (e Event) MarshalData() ([]byte, error) {
	return json.Marshal(e.Data)
}