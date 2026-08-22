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

// NewCollector creates a new telemetry flow collector with pre-seeded demonstration traffic.
func NewCollector(log *zap.Logger) *Collector {
	c := &Collector{
		flows:  make([]NetworkFlow, 0, 100),
		maxCap: 500,
		log:    log.Named("hubble-collector"),
	}
	c.seedInitialFlows()
	return c
}

func (c *Collector) seedInitialFlows() {
	now := time.Now()
	sampleFlows := []NetworkFlow{
		{
			ID:         "flow-1",
			Timestamp:  now.Add(-2 * time.Second),
			SrcPod:     "frontend-ingress-7d4a",
			SrcNS:      "default",
			SrcIP:      "10.244.0.12",
			DstPod:     "api-gateway-55b2",
			DstNS:      "default",
			DstIP:      "10.244.0.15",
			DstPort:    8080,
			Protocol:   "HTTP",
			Verdict:    "FORWARDED",
			StatusCode: 200,
			LatencyMs:  1.4,
			Bytes:      2048,
			Summary:    "GET /api/v1/users (200 OK)",
		},
		{
			ID:         "flow-2",
			Timestamp:  now.Add(-1 * time.Second),
			SrcPod:     "api-gateway-55b2",
			SrcNS:      "default",
			SrcIP:      "10.244.0.15",
			DstPod:     "user-auth-service-91a",
			DstNS:      "default",
			DstIP:      "10.244.0.18",
			DstPort:    50051,
			Protocol:   "TCP",
			Verdict:    "FORWARDED",
			StatusCode: 200,
			LatencyMs:  0.8,
			Bytes:      1024,
			Summary:    "gRPC VerifySession",
		},
		{
			ID:         "flow-3",
			Timestamp:  now.Add(-500 * time.Millisecond),
			SrcPod:     "unknown-scanner-pod",
			SrcNS:      "tarak-public",
			SrcIP:      "192.168.1.105",
			DstPod:     "db-primary-0",
			DstNS:      "tarak-system",
			DstIP:      "10.244.0.5",
			DstPort:    5432,
			Protocol:   "TCP",
			Verdict:    "DROPPED",
			StatusCode: 403,
			LatencyMs:  0.1,
			Bytes:      64,
			Summary:    "Strict NetworkPolicy Blocked",
		},
	}
	c.flows = append(c.flows, sampleFlows...)
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
