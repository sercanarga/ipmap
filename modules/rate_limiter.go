package modules

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	rate       int        // requests per second
	tokens     int        // current available tokens
	maxTokens  int        // maximum tokens (burst size)
	lastRefill time.Time  // last time tokens were refilled
	mu         sync.Mutex // mutex for thread safety
	enabled    bool       // whether rate limiting is enabled
}

// NewRateLimiter creates a new rate limiter
// rate: requests per second (0 = disabled)
// burst: maximum burst size (defaults to rate if 0)
func NewRateLimiter(rate int, burst int) *RateLimiter {
	if rate <= 0 {
		return &RateLimiter{enabled: false}
	}

	if burst <= 0 {
		burst = rate
	}

	return &RateLimiter{
		rate:       rate,
		tokens:     burst,
		maxTokens:  burst,
		lastRefill: time.Now(),
		enabled:    true,
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	if !rl.enabled {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	for rl.tokens <= 0 {
		rl.mu.Unlock()
		// Wait for next token
		time.Sleep(time.Second / time.Duration(rl.rate))
		rl.mu.Lock()
		rl.refillTokens()
	}

	rl.tokens--
}

// TryAcquire attempts to get a token without blocking
// Returns true if successful, false otherwise
func (rl *RateLimiter) TryAcquire() bool {
	if !rl.enabled {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// refillTokens adds tokens based on elapsed time (must be called with lock held)
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// Calculate tokens to add based on elapsed time
	tokensToAdd := int(elapsed.Seconds() * float64(rl.rate))

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}
}

// IsEnabled returns whether rate limiting is enabled
func (rl *RateLimiter) IsEnabled() bool {
	return rl.enabled
}

// GetRate returns the current rate limit
func (rl *RateLimiter) GetRate() int {
	return rl.rate
}

// SetRate updates the rate limit
func (rl *RateLimiter) SetRate(rate int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rate <= 0 {
		rl.enabled = false
		return
	}

	rl.rate = rate
	rl.maxTokens = rate
	rl.enabled = true
}
