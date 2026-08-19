// Package metrics provides Prometheus-compatible metrics for the Tarak API server.
//
// Metrics are registered in a private Prometheus registry (not the default global
// registry) to allow precise control and avoid pollution from transitive dependencies.
//
// Exposed at: GET /metrics (Prometheus text format)
//
// Metric naming follows Prometheus conventions:
//
//	tarak_apiserver_*    — API server metrics
//	tarak_statestore_*   — State store metrics
//	tarak_scheduler_*    — Scheduler metrics (Phase 3)
//	tarak_controller_*   — Controller metrics (Phase 4)
//	tarak_agent_*        — Node agent metrics (Phase 2)
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry is the central metrics registry for a Tarak process.
type Registry struct {
	reg *prometheus.Registry

	// ─── API server metrics ──────────────────────────────────────────────────

	// APIRequestsTotal counts all API requests by method, resource, and HTTP status.
	APIRequestsTotal *prometheus.CounterVec

	// APIRequestDuration measures API handler latency.
	APIRequestDuration *prometheus.HistogramVec

	// APIWatchersActive is the current number of active watch streams.
	APIWatchersActive prometheus.Gauge

	// ─── State store metrics ──────────────────────────────────────────────────

	// StateStoreOperationsTotal counts state store operations by type (create/update/delete/get/list).
	StateStoreOperationsTotal *prometheus.CounterVec

	// StateStoreOperationDuration measures state store operation latency.
	StateStoreOperationDuration *prometheus.HistogramVec

	// StateStoreRevision is the current global revision number.
	StateStoreRevision prometheus.Gauge

	// StateStoreObjectsTotal is the total number of stored objects by resource type.
	StateStoreObjectsTotal *prometheus.GaugeVec

	// ─── Auth metrics ─────────────────────────────────────────────────────────

	// AuthRequestsTotal counts auth attempts by method (cert/token) and result.
	AuthRequestsTotal *prometheus.CounterVec
}

// NewRegistry creates a new metrics Registry with all metrics registered.
// Call Register() on the returned value to add it to the global Prometheus default registry,
// or use its Handler() for a custom /metrics endpoint.
func NewRegistry() *Registry {
	reg := prometheus.NewRegistry()

	// Add Go runtime and process metrics.
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	r := &Registry{reg: reg}

	r.APIRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tarak",
		Subsystem: "apiserver",
		Name:      "requests_total",
		Help:      "Total number of API requests by method, resource, verb, and HTTP status code.",
	}, []string{"method", "resource", "verb", "code"})

	r.APIRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "tarak",
		Subsystem: "apiserver",
		Name:      "request_duration_seconds",
		Help:      "Latency of API server requests in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "resource", "verb"})

	r.APIWatchersActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "tarak",
		Subsystem: "apiserver",
		Name:      "watchers_active",
		Help:      "Current number of active watch stream connections.",
	})

	r.StateStoreOperationsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tarak",
		Subsystem: "statestore",
		Name:      "operations_total",
		Help:      "Total number of state store operations by type and result.",
	}, []string{"operation", "result"})

	r.StateStoreOperationDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "tarak",
		Subsystem: "statestore",
		Name:      "operation_duration_seconds",
		Help:      "Latency of state store operations.",
		Buckets:   []float64{.0001, .0005, .001, .005, .01, .05, .1, .5, 1, 5},
	}, []string{"operation"})

	r.StateStoreRevision = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "tarak",
		Subsystem: "statestore",
		Name:      "current_revision",
		Help:      "Current global revision number of the state store.",
	})

	r.StateStoreObjectsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "tarak",
		Subsystem: "statestore",
		Name:      "objects_total",
		Help:      "Total number of stored objects by resource group/version/kind.",
	}, []string{"group", "version", "resource"})

	r.AuthRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tarak",
		Subsystem: "apiserver",
		Name:      "auth_requests_total",
		Help:      "Total number of authentication attempts by method and result.",
	}, []string{"method", "result"})

	reg.MustRegister(
		r.APIRequestsTotal,
		r.APIRequestDuration,
		r.APIWatchersActive,
		r.StateStoreOperationsTotal,
		r.StateStoreOperationDuration,
		r.StateStoreRevision,
		r.StateStoreObjectsTotal,
		r.AuthRequestsTotal,
	)

	return r
}

// Handler returns an HTTP handler that serves Prometheus metrics.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{
		EnableOpenMetrics: false,
	})
}

// MustRegister registers one or more collectors with the registry.
func (r *Registry) MustRegister(cs ...prometheus.Collector) {
	r.reg.MustRegister(cs...)
}
