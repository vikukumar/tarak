// Package statestore — Watch event fan-out bus.
//
// The bus receives write events from the BoltDB store and fans them out to all
// registered watchers. Each watcher has its own buffered channel; slow watchers
// are dropped (with an error logged) rather than blocking the store.
//
// Historical event replay is backed by a bounded in-memory ring buffer that holds
// the last RingBufferSize events. Callers requesting history beyond the ring buffer
// receive events starting from the oldest available revision.
package statestore

import (
	"context"
	"sync"
)

const (
	// defaultWatcherBufferSize is the number of events each watcher channel can buffer
	// before the watcher is considered overloaded and dropped.
	defaultWatcherBufferSize = 256

	// ringBufferSize is the number of recent events retained for watch replay.
	// Clients requesting a sinceRevision older than (currentRevision - ringBufferSize)
	// will miss events and should re-list.
	ringBufferSize = 65536
)

// ─── watchBus ─────────────────────────────────────────────────────────────────

// watchBus is the central event distribution hub.
// It is goroutine-safe and designed for high-throughput fan-out.
type watchBus struct {
	mu      sync.RWMutex
	subs    map[uint64]*watchSubscriber
	nextID  uint64

	// ring is a circular buffer of recent WatchEvents for history replay.
	ring     [ringBufferSize]storedEvent
	ringHead int64 // points to the slot of the oldest event
	ringSize int64 // number of events currently in the ring

	// currentRevision is the latest revision known to the bus.
	currentRevision int64
}

// storedEvent is what we store in the ring buffer.
type storedEvent struct {
	revision int64
	event    WatchEvent
}

// watchSubscriber is a single registered watcher.
type watchSubscriber struct {
	id     uint64
	query  WatchQuery
	ch     chan WatchEvent
	once   sync.Once // ensures close(ch) is called exactly once
}

// newWatchBus constructs an initialised watchBus.
func newWatchBus() *watchBus {
	return &watchBus{
		subs: make(map[uint64]*watchSubscriber),
	}
}

// subscribe registers a new watcher and returns its event channel.
// The channel is closed when the context is cancelled or the watcher is dropped.
func (b *watchBus) subscribe(ctx context.Context, query WatchQuery) (<-chan WatchEvent, error) {
	b.mu.Lock()

	id := b.nextID
	b.nextID++

	sub := &watchSubscriber{
		id:    id,
		query: query,
		ch:    make(chan WatchEvent, defaultWatcherBufferSize),
	}
	b.subs[id] = sub

	// Capture ring state while holding the write lock so we can replay atomically.
	var history []WatchEvent
	if query.SinceRevision > 0 {
		history = b.historyLocked(query)
	}
	currentRev := b.currentRevision

	b.mu.Unlock()

	// Replay historical events before releasing the lock so the subscriber sees
	// a consistent, gap-free stream.
	for _, ev := range history {
		select {
		case sub.ch <- ev:
		default:
			// Buffer full during replay — unsubscribe and return error.
			b.unsubscribe(id)
			return nil, ErrWatcherCapacityExceeded
		}
	}

	// Send initial BOOKMARK so the client knows where history ends.
	if query.SinceRevision > 0 || query.SendBookmarks {
		bm := WatchEvent{
			Type: EventBookmark,
			Envelope: &Envelope{ResourceVersion: currentRev},
			Key:  query.Key,
		}
		select {
		case sub.ch <- bm:
		default:
			b.unsubscribe(id)
			return nil, ErrWatcherCapacityExceeded
		}
	}

	// Background goroutine to close the channel when the context is cancelled.
	go func() {
		<-ctx.Done()
		b.unsubscribe(id)
	}()

	return sub.ch, nil
}

// unsubscribe removes a watcher by ID and closes its channel.
func (b *watchBus) unsubscribe(id uint64) {
	b.mu.Lock()
	sub, ok := b.subs[id]
	if ok {
		delete(b.subs, id)
	}
	b.mu.Unlock()

	if ok {
		sub.once.Do(func() { close(sub.ch) })
	}
}

// publish broadcasts a WatchEvent to all matching subscribers and appends it
// to the ring buffer for future history requests.
//
// publish is called from the BoltDB write path and must not block.
// Subscribers whose channels are full are dropped.
func (b *watchBus) publish(ev WatchEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Update ring buffer.
	slot := b.currentRevision % ringBufferSize
	b.ring[slot] = storedEvent{revision: ev.Envelope.ResourceVersion, event: ev}
	b.currentRevision = ev.Envelope.ResourceVersion
	if b.ringSize < ringBufferSize {
		b.ringSize++
	} else {
		// Ring is full; advance head (oldest entry is overwritten).
		b.ringHead = (b.ringHead + 1) % ringBufferSize
	}

	// Fan-out to matching subscribers.
	var toRemove []uint64
	for id, sub := range b.subs {
		if !b.matchesSubscriber(sub, ev) {
			continue
		}
		select {
		case sub.ch <- ev:
		default:
			// Subscriber is slow — remove it.
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		sub := b.subs[id]
		delete(b.subs, id)
		sub.once.Do(func() { close(sub.ch) })
	}
}

// matchesSubscriber reports whether ev should be delivered to sub.
// Must be called with b.mu held.
func (b *watchBus) matchesSubscriber(sub *watchSubscriber, ev WatchEvent) bool {
	q := sub.query
	k := ev.Key

	// Resource type must match.
	if q.Key.Group != k.Group || q.Key.Version != k.Version || q.Key.Resource != k.Resource {
		return false
	}
	// Namespace filter.
	if q.Key.Namespace != "" && q.Key.Namespace != k.Namespace {
		return false
	}
	// Exact name filter.
	if q.Key.Name != "" && q.Key.Name != k.Name {
		return false
	}
	// Label selector.
	if q.LabelSelector != nil && ev.Envelope != nil {
		if !q.LabelSelector.Matches(ev.Envelope.Labels) {
			return false
		}
	}
	return true
}

// historyLocked replays events from the ring buffer that are newer than
// query.SinceRevision and match the query.  Must be called with b.mu held.
func (b *watchBus) historyLocked(query WatchQuery) []WatchEvent {
	if b.ringSize == 0 {
		return nil
	}
	var events []WatchEvent
	for i := int64(0); i < b.ringSize; i++ {
		idx := (b.ringHead + i) % ringBufferSize
		se := b.ring[idx]
		if se.revision <= query.SinceRevision {
			continue
		}
		if !b.matchesSubscriber(&watchSubscriber{query: query}, se.event) {
			continue
		}
		events = append(events, se.event)
	}
	return events
}

// subscriberCount returns the current number of active watchers.
// Used for testing and metrics.
func (b *watchBus) subscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}
