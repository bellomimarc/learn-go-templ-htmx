package dashboard

import (
	"sync"
	"time"
)

// RateLimiter enforces a request limit per IP address within a time window.
// It is thread-safe and can be reused across multiple endpoints.
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	maxReqs  int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter with the given constraints.
// maxReqs: maximum number of requests allowed
// window: time window to consider for the limit
func NewRateLimiter(maxReqs int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		maxReqs:  maxReqs,
		window:   window,
	}
}

// IsAllowed checks if a request from the given IP is allowed.
// Returns true if allowed, false if rate limited.
func (rl *RateLimiter) IsAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get or initialize request list for this IP
	reqs, exists := rl.requests[ip]
	if !exists {
		rl.requests[ip] = []time.Time{now}
		return true
	}

	// Remove old requests outside the window
	var recent []time.Time
	for _, t := range reqs {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	// Check if limit exceeded
	if len(recent) >= rl.maxReqs {
		return false // Rate limited
	}

	// Add current request
	recent = append(recent, now)
	rl.requests[ip] = recent
	return true // Allowed
}
