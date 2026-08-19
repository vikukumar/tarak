// Package statestore — Store interface and error types.
package statestore

import (
	"context"
	"errors"
	"fmt"
)

// ─── Store ────────────────────────────────────────────────────────────────────

// Store is the Tarak state store interface.
//
// All methods are safe for concurrent use from multiple goroutines.
// Implementations must:
//   - Use optimistic concurrency on Update / Delete.
//   - Increment the global revision on every write.
//   - Deliver watch events reliably to all registered watchers.
//   - Support graceful shutdown via context cancellation.
type Store interface {
	// Create persists a new object.
	// Returns ErrAlreadyExists if an object with the same key already exists.
	// Returns ErrKeyInvalid if the key does not have all required fields.
	// On success the returned Envelope has ResourceVersion set by the store.
	Create(ctx context.Context, key ResourceKey, obj RawObject) (*Envelope, error)

	// Update replaces an existing object.
	// resourceVersion must match the currently stored value (optimistic concurrency).
	// Returns ErrNotFound if the object does not exist.
	// Returns ErrConflict if the resourceVersion does not match.
	// On success the returned Envelope has the new ResourceVersion.
	Update(ctx context.Context, key ResourceKey, obj RawObject, resourceVersion int64) (*Envelope, error)

	// StatusUpdate updates only the status subresource of an object.
	// This increments resourceVersion but does NOT increment generation.
	StatusUpdate(ctx context.Context, key ResourceKey, obj RawObject, resourceVersion int64) (*Envelope, error)

	// Delete removes an object.
	// If resourceVersion > 0, it must match (optimistic concurrency).
	// Pass resourceVersion = 0 to force-delete regardless of version.
	// Returns ErrNotFound if the object does not exist.
	Delete(ctx context.Context, key ResourceKey, resourceVersion int64) (*Envelope, error)

	// Get retrieves a single object.
	// Returns ErrNotFound if the object does not exist.
	Get(ctx context.Context, key ResourceKey) (*Envelope, error)

	// List retrieves all objects matching the query.
	// Returns the list and the current global revision (for use as a watch starting point).
	List(ctx context.Context, query ListQuery) ([]*Envelope, int64, error)

	// Watch returns a channel of events starting from the given revision.
	// If sinceRevision is 0, only future events are delivered.
	// If sinceRevision > 0, historical events since that revision are replayed first (bounded).
	// The channel is closed when the context is cancelled or the watcher is dropped.
	Watch(ctx context.Context, query WatchQuery) (<-chan WatchEvent, error)

	// CurrentRevision returns the current global revision number.
	CurrentRevision(ctx context.Context) (int64, error)

	// Close closes the store, flushing pending writes.
	// After Close returns, no more writes are accepted.
	Close() error
}

// ─── Query Types ──────────────────────────────────────────────────────────────

// ListQuery specifies which objects to return from a List call.
type ListQuery struct {
	// Key specifies the resource type and optional namespace.
	// Key.Name is ignored for list queries.
	Key ResourceKey

	// LabelSelector filters by labels.  nil = match all.
	LabelSelector ParsedLabelSelector

	// FieldSelector filters by fields.  nil = match all.
	FieldSelector ParsedFieldSelector

	// Limit is the maximum number of results.  0 = no limit.
	Limit int64

	// Continue is an opaque token returned from a previous list for pagination.
	Continue string

	// ResourceVersion constrains which version of the list is returned.
	// 0 = return current state.
	ResourceVersion int64
}

// WatchQuery specifies the scope of a Watch.
type WatchQuery struct {
	// Key specifies the resource type and optional namespace / name.
	// If Key.Name is set, only events for that specific object are delivered.
	// If Key.Namespace is set (without Name), only events in that namespace are delivered.
	Key ResourceKey

	// LabelSelector filters delivered events.  nil = match all.
	LabelSelector ParsedLabelSelector

	// FieldSelector filters delivered events.  nil = match all.
	FieldSelector ParsedFieldSelector

	// SinceRevision is the resource version from which to replay history.
	// 0 = only future events.
	SinceRevision int64

	// SendBookmarks requests periodic BOOKMARK events to advance the watch cursor.
	SendBookmarks bool
}

// WatchEvent is a single event delivered on a watch channel.
type WatchEvent struct {
	// Type is ADDED, MODIFIED, DELETED, or BOOKMARK.
	Type EventType
	// Envelope is the full object state at the time of the event.
	// For DELETED events, it contains the object's last known state.
	// For BOOKMARK events, only ResourceVersion is meaningful.
	Envelope *Envelope
	// Key identifies the resource this event is about.
	Key ResourceKey
}

// EventType describes the type of a watch event.
type EventType string

const (
	EventAdded    EventType = "ADDED"
	EventModified EventType = "MODIFIED"
	EventDeleted  EventType = "DELETED"
	EventBookmark EventType = "BOOKMARK"
	EventError    EventType = "ERROR"
)

// RawObject is the raw JSON of a resource as provided by the caller.
type RawObject = []byte

// ParsedLabelSelector is imported from the meta package selectors.
// Re-declared here to avoid an import cycle.
type ParsedLabelSelector = interface {
	Matches(labels map[string]string) bool
	Empty() bool
}

// ParsedFieldSelector is imported from the meta package selectors.
type ParsedFieldSelector = interface {
	Matches(fields map[string]string) bool
	Empty() bool
}

// ─── Sentinel Errors ──────────────────────────────────────────────────────────

// Sentinel error values.  Use errors.Is() to check.
var (
	// ErrNotFound is returned when the requested object does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned when creating an object that already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrConflict is returned when the provided resourceVersion does not match
	// the stored version (optimistic concurrency violation).
	ErrConflict = errors.New("resource version conflict")

	// ErrKeyInvalid is returned when a ResourceKey does not have required fields.
	ErrKeyInvalid = errors.New("invalid resource key")

	// ErrStoreClosed is returned when an operation is attempted on a closed store.
	ErrStoreClosed = errors.New("store is closed")

	// ErrWatcherCapacityExceeded is returned when the watcher's event channel is full
	// and the watcher is dropped.
	ErrWatcherCapacityExceeded = errors.New("watcher capacity exceeded, watcher dropped")
)

// ─── Error Helpers ────────────────────────────────────────────────────────────

// NotFoundError wraps ErrNotFound with the resource key that was missing.
type NotFoundError struct {
	Key ResourceKey
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource not found: %s", e.Key)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// AlreadyExistsError wraps ErrAlreadyExists with the resource key.
type AlreadyExistsError struct {
	Key ResourceKey
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("resource already exists: %s", e.Key)
}

func (e *AlreadyExistsError) Is(target error) bool {
	return target == ErrAlreadyExists
}

// ConflictError wraps ErrConflict with the resource key and version info.
type ConflictError struct {
	Key              ResourceKey
	ProvidedVersion  int64
	CurrentVersion   int64
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("resource conflict: %s: provided version %d does not match current version %d",
		e.Key, e.ProvidedVersion, e.CurrentVersion)
}

func (e *ConflictError) Is(target error) bool {
	return target == ErrConflict
}
