package mesh

import (
	"time"
)

// RateLimitPolicy defines HTTP and L4 request throttling rules.
type RateLimitPolicy struct {
	Name        string   `json:"name"`
	Mesh        string   `json:"mesh"`
	Namespace   string   `json:"namespace"`
	TargetRef   string   `json:"targetRef"`   // Target service e.g. "orders-service"
	RequestsPerSec int   `json:"requestsPerSec"`
	BurstSize   int      `json:"burstSize"`
	Unit        string   `json:"unit"`        // "second", "minute", "hour"
	Enabled     bool     `json:"enabled"`
}

// CircuitBreakerPolicy defines fail-fast protection to isolate failing instances.
type CircuitBreakerPolicy struct {
	Name             string `json:"name"`
	Mesh             string `json:"mesh"`
	Namespace        string `json:"namespace"`
	TargetRef        string `json:"targetRef"`
	MaxConnections   int    `json:"maxConnections"`
	MaxPendingReqs   int    `json:"maxPendingReqs"`
	ConsecutiveErrors int   `json:"consecutiveErrors"`
	IntervalSeconds  int    `json:"intervalSeconds"`
	BaseEjectionTime string `json:"baseEjectionTime"` // e.g. "30s"
	Enabled          bool   `json:"enabled"`
}

// FaultInjectionPolicy simulates real-world failure modes for chaos testing.
type FaultInjectionPolicy struct {
	Name        string `json:"name"`
	Mesh        string `json:"mesh"`
	Namespace   string `json:"namespace"`
	TargetRef   string `json:"targetRef"`
	DelayType   string `json:"delayType"`   // "fixed"
	DelayValue  string `json:"delayValue"`  // "200ms", "2s"
	DelayPercent int   `json:"delayPercent"`
	AbortCode   int    `json:"abortCode"`   // 503, 500
	AbortPercent int   `json:"abortPercent"`
	Enabled     bool   `json:"enabled"`
}

// HealthCheckPolicy configures active end-to-end probing for upstream mesh workloads.
type HealthCheckPolicy struct {
	Name               string        `json:"name"`
	Mesh               string        `json:"mesh"`
	Namespace          string        `json:"namespace"`
	TargetRef          string        `json:"targetRef"`
	Protocol           string        `json:"protocol"` // "http", "tcp", "grpc"
	Path               string        `json:"path"`     // "/healthz"
	Interval           time.Duration `json:"interval"`
	Timeout            time.Duration `json:"timeout"`
	HealthyThreshold   int           `json:"healthyThreshold"`
	UnhealthyThreshold int           `json:"unhealthyThreshold"`
	Enabled            bool          `json:"enabled"`
}
