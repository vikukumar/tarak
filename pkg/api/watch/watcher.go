// Package watch implements the SSE (Server-Sent Events) watch stream handler
// for the Tarak API server.
//
// Kubernetes uses HTTP chunked streaming for watches.  Each event is written as
// a complete JSON object followed by a newline.  Clients like kubectl parse
// newline-delimited JSON from the response body.
//
// Request:  GET /api/v1/namespaces/default/pods?watch=true&resourceVersion=42
// Response: HTTP 200 with Transfer-Encoding: chunked
//           {"type":"ADDED","object":{...}}
//           {"type":"MODIFIED","object":{...}}
//           …  (stream stays open until client disconnects or timeout)
package watch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/statestore"
	"github.com/vikukumar/tarak/api/meta"
)

// WireEvent is the JSON structure written on the wire for each watch event.
// It matches the Kubernetes watch event format.
type WireEvent struct {
	Type   string          `json:"type"`
	Object json.RawMessage `json:"object"`
}

// Options configures a Watch handler.
type Options struct {
	// Store is the state store from which events are sourced.
	Store statestore.Store
	// Log is the structured logger.
	Log *zap.Logger
	// DefaultTimeout is the maximum duration of a watch connection (default 5m).
	// Clients can override with ?timeoutSeconds=N.
	DefaultTimeout time.Duration
}

// Handler is the watch stream HTTP handler.
type Handler struct {
	store  statestore.Store
	log    *zap.Logger
	defTO  time.Duration
}

// New creates a new watch Handler.
func New(opts Options) *Handler {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}
	if opts.DefaultTimeout == 0 {
		opts.DefaultTimeout = 5 * time.Minute
	}
	return &Handler{
		store: opts.Store,
		log:   opts.Log,
		defTO: opts.DefaultTimeout,
	}
}

// ServeWatch handles a watch request.
//
// Parameters extracted from the request:
//   - query.Key — resource type and optional namespace/name
//   - sinceRevision — from ?resourceVersion=N
//   - timeout — from ?timeoutSeconds=N (default Options.DefaultTimeout)
//   - labelSelector — from ?labelSelector=…
//   - fieldSelector — from ?fieldSelector=…
func (h *Handler) ServeWatch(
	w http.ResponseWriter,
	r *http.Request,
	query statestore.WatchQuery,
) {
	// Set up the response for streaming.
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported by this transport", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Determine timeout.
	timeout := h.defTO
	if ts := r.URL.Query().Get("timeoutSeconds"); ts != "" {
		var secs int
		if _, err := fmt.Sscanf(ts, "%d", &secs); err == nil && secs > 0 {
			timeout = time.Duration(secs) * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	// Subscribe to the watch bus.
	ch, err := h.store.Watch(ctx, query)
	if err != nil {
		h.log.Warn("watch subscribe failed", zap.Error(err))
		// Write an error event.
		writeErrorEvent(w, flusher, err)
		return
	}

	h.log.Debug("watch started",
		zap.String("resource", query.Key.Resource),
		zap.String("namespace", query.Key.Namespace),
		zap.String("name", query.Key.Name),
		zap.Int64("sinceRevision", query.SinceRevision),
	)

	enc := json.NewEncoder(w)

	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				// Channel closed — client context cancelled or watcher dropped.
				return
			}
			wireEv := WireEvent{
				Type:   string(ev.Type),
				Object: ev.Envelope.Object,
			}
			if wireEv.Object == nil {
				// BOOKMARK event has no full object.
				rv := fmt.Sprintf("%d", ev.Envelope.ResourceVersion)
				bm := struct {
					meta.TypeMeta   `json:",inline"`
					meta.ObjectMeta `json:"metadata"`
				}{}
				bm.TypeMeta.Kind = "Event"
				bm.TypeMeta.APIVersion = "v1"
				bm.ObjectMeta.ResourceVersion = rv
				bm.ObjectMeta.Annotations = map[string]string{
					"tarak.io/watch-bookmark-revision": rv,
				}
				obj, _ := json.Marshal(bm)
				wireEv.Object = obj
			}
			if err := enc.Encode(wireEv); err != nil {
				h.log.Debug("watch encode error", zap.Error(err))
				return
			}
			flusher.Flush()

		case <-ctx.Done():
			return
		}
	}
}

// writeErrorEvent writes a single ERROR watch event to the stream.
func writeErrorEvent(w http.ResponseWriter, f http.Flusher, err error) {
	errStatus := struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Status     string `json:"status"`
		Message    string `json:"message"`
		Code       int    `json:"code"`
	}{
		APIVersion: "v1",
		Kind:       "Status",
		Status:     "Failure",
		Message:    err.Error(),
		Code:       http.StatusInternalServerError,
	}
	errBytes, _ := json.Marshal(errStatus)
	wireEv := WireEvent{
		Type:   string(statestore.EventError),
		Object: errBytes,
	}
	data, _ := json.Marshal(wireEv)
	_, _ = fmt.Fprintf(w, "%s\n", data)
	f.Flush()
}
