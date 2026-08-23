package mesh

import (
	"sync"
	"time"
)

type tokenBucket struct {
	tokens     float64
	capacity   float64
	rate       float64 // tokens per second
	lastRefill time.Time
}

// RateLimiter manages dynamic token-bucket rate limits per client key.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

// NewRateLimiter creates a RateLimiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*tokenBucket),
	}
}

// Allow checks if a request is permitted under capacity and rate.
func (rl *RateLimiter) Allow(key string, capacity float64, rate float64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &tokenBucket{
			tokens:     capacity - 1,
			capacity:   capacity,
			rate:       rate,
			lastRefill: now,
		}
		rl.buckets[key] = b
		return true
	}

	// Refill tokens
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = b.tokens + elapsed*b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}

	return false
}
