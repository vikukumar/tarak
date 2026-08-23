package mesh

import (
	"sync"
	"time"
)

// CircuitBreaker manages outlier detection and failover for a service backend.
type CircuitBreaker struct {
	mu               sync.Mutex
	MaxFailures      int
	CooldownDuration time.Duration
	failureCount     map[string]int
	trippedUntil     map[string]time.Time
}

// NewCircuitBreaker creates a circuit breaker with failure thresholds.
func NewCircuitBreaker(maxFailures int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		MaxFailures:      maxFailures,
		CooldownDuration: cooldown,
		failureCount:     make(map[string]int),
		trippedUntil:     make(map[string]time.Time),
	}
}

// Allow checks if the destination host is healthy or currently tripped.
func (cb *CircuitBreaker) Allow(host string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	until, tripped := cb.trippedUntil[host]
	if !tripped {
		return true
	}

	if time.Now().After(until) {
		// Cooldown expired, enter half-open state
		delete(cb.trippedUntil, host)
		cb.failureCount[host] = 0
		return true
	}

	return false
}

// ReportFailure increments failure count and trips breaker if threshold is reached.
func (cb *CircuitBreaker) ReportFailure(host string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount[host]++
	if cb.failureCount[host] >= cb.MaxFailures {
		cb.trippedUntil[host] = time.Now().Add(cb.CooldownDuration)
	}
}

// ReportSuccess resets the failure counter.
func (cb *CircuitBreaker) ReportSuccess(host string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount[host] = 0
	delete(cb.trippedUntil, host)
}
