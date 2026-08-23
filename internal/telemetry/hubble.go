package telemetry

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// NetworkFlow represents a single network packet or connection event.
type NetworkFlow struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	SrcPod      string    `json:"srcPod"`
	SrcNS       string    `json:"srcNs"`
	SrcIP       string    `json:"srcIp"`
	DstPod      string    `json:"dstPod"`
	DstNS       string    `json:"dstNs"`
	DstIP       string    `json:"dstIp"`
	DstPort     int       `json:"dstPort"`
	Protocol    string    `json:"protocol"` // "HTTP", "HTTPS", "TCP", "UDP", "DNS"
	Verdict     string    `json:"verdict"`  // "FORWARDED", "DROPPED", "AUDIT"
	StatusCode  int       `json:"statusCode,omitempty"`
	LatencyMs   float64   `json:"latencyMs"`
	Bytes       int64     `json:"bytes"`
	Summary     string    `json:"summary"`
}

// Collector records and serves Hubble-style network flow telemetry.
type Collector struct {
	mu     sync.RWMutex
	flows  []NetworkFlow
	maxCap int
	log    *zap.Logger
}

// NewCollector creates a new telemetry flow collector.
func NewCollector(log *zap.Logger) *Collector {
	return &Collector{
		flows:  make([]NetworkFlow, 0, 100),
		maxCap: 500,
		log:    log.Named("hubble-collector"),
	}
}

// RecordFlow appends a network flow event.
func (c *Collector) RecordFlow(f NetworkFlow) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.flows) >= c.maxCap {
		c.flows = c.flows[1:]
	}
	c.flows = append(c.flows, f)
}

// ListFlows returns recent network flows.
func (c *Collector) ListFlows() []NetworkFlow {
	c.mu.RLock()
	defer c.mu.RUnlock()

	copied := make([]NetworkFlow, len(c.flows))
	copy(copied, c.flows)
	return copied
}

// HandleListFlows serves GET /apis/telemetry.tarak.io/v1/flows.
func (c *Collector) HandleListFlows(w http.ResponseWriter, r *http.Request) {
	flows := c.ListFlows()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"apiVersion": "telemetry.tarak.io/v1",
		"kind":       "FlowList",
		"items":      flows,
	})
}
