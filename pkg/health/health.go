// Package health implements the Tarak API server health check endpoints.
//
// Three endpoints are exposed, matching the Kubernetes convention:
//
//	GET /healthz   — basic liveness; always returns 200 if the process is up
//	GET /readyz    — readiness; returns 200 only when the server is fully initialised
//	GET /livez     — liveness (alias for /healthz with structured output)
//
// Each endpoint supports an optional ?verbose=true query parameter that includes
// the status of individual health checks in the response body.
package health

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
)

// ─── Checker ──────────────────────────────────────────────────────────────────

// CheckFunc is a function that reports the health of a subsystem.
// It returns nil if healthy, or a non-nil error describing the problem.
type CheckFunc func() error

// Handler is the health check HTTP handler.
// It registers named health checks and serves /healthz, /readyz, /livez.
type Handler struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
	ready  atomic.Bool
}

// NewHandler creates a new Handler with no registered checks.
func NewHandler() *Handler {
	return &Handler{
		checks: make(map[string]CheckFunc),
	}
}

// AddCheck registers a named health check.
// Checks are evaluated in registration order during every health request.
// Thread-safe; can be called after the server has started.
func (h *Handler) AddCheck(name string, fn CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = fn
}

// SetReady marks the server as ready to serve traffic.
// Until SetReady is called, /readyz returns 503.
func (h *Handler) SetReady(ready bool) {
	h.ready.Store(ready)
}

// ─── HTTP Handlers ────────────────────────────────────────────────────────────

// Healthz handles GET /healthz.
// Always returns 200 if the process is up (liveness).
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, false)
}

// Readyz handles GET /readyz.
// Returns 503 until SetReady(true) has been called, then runs all checks.
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	if !h.ready.Load() {
		http.Error(w, `{"status":"not ready"}`, http.StatusServiceUnavailable)
		return
	}
	h.respond(w, r, true)
}

// Livez handles GET /livez.
// Identical to Healthz but with a different path for Kubernetes compatibility.
func (h *Handler) Livez(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, false)
}

// respond evaluates health checks and writes the response.
func (h *Handler) respond(w http.ResponseWriter, r *http.Request, runChecks bool) {
	verbose := r.URL.Query().Get("verbose") == "true"

	if !runChecks && !verbose {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}

	h.mu.RLock()
	checks := make(map[string]CheckFunc, len(h.checks))
	for k, v := range h.checks {
		checks[k] = v
	}
	h.mu.RUnlock()

	type checkResult struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}

	var results []checkResult
	allOK := true

	for name, fn := range checks {
		res := checkResult{Name: name, Status: "ok"}
		if err := fn(); err != nil {
			res.Status = "error"
			res.Message = err.Error()
			allOK = false
		}
		results = append(results, res)
	}

	type response struct {
		Status string        `json:"status"`
		Checks []checkResult `json:"checks,omitempty"`
	}

	resp := response{
		Status: "ok",
	}
	if !allOK {
		resp.Status = "error"
	}
	if verbose {
		resp.Checks = results
	}

	w.Header().Set("Content-Type", "application/json")
	if !allOK {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// chiRouter is the minimal interface needed to register routes.
// This matches chi.Router and http.ServeMux (via HandleFunc).
type chiRouter interface {
	Get(pattern string, handlerFn http.HandlerFunc)
}

// RegisterRoutes attaches the health check handlers to a chi-compatible router.
func (h *Handler) RegisterRoutes(r chiRouter) {
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)
	r.Get("/livez", h.Livez)
}
