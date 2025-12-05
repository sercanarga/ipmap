package modules

import (
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name    string
		rate    int
		burst   int
		enabled bool
	}{
		{"Disabled (rate 0)", 0, 0, false},
		{"Disabled (negative)", -1, 0, false},
		{"Enabled", 10, 5, true},
		{"Default burst", 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rate, tt.burst)
			if rl.IsEnabled() != tt.enabled {
				t.Errorf("NewRateLimiter(%d, %d).IsEnabled() = %v, want %v",
					tt.rate, tt.burst, rl.IsEnabled(), tt.enabled)
			}
		})
	}
}

func TestRateLimiterWait(t *testing.T) {
	// Test that disabled limiter doesn't block
	rl := NewRateLimiter(0, 0)
	start := time.Now()
	for i := 0; i < 10; i++ {
		rl.Wait()
	}
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Disabled rate limiter should not block, took %v", elapsed)
	}
}

func TestRateLimiterTryAcquire(t *testing.T) {
	rl := NewRateLimiter(10, 5)

	// Should be able to acquire burst tokens immediately
	for i := 0; i < 5; i++ {
		if !rl.TryAcquire() {
			t.Errorf("TryAcquire() should succeed for token %d within burst", i+1)
		}
	}

	// Next one should fail (no tokens left)
	if rl.TryAcquire() {
		t.Error("TryAcquire() should fail after burst exhausted")
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := NewRateLimiter(100, 10)
	var wg sync.WaitGroup
	acquired := 0
	var mu sync.Mutex

	// Try to acquire from multiple goroutines
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.TryAcquire() {
				mu.Lock()
				acquired++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should have acquired exactly 10 (burst size)
	if acquired != 10 {
		t.Errorf("Should have acquired 10 tokens, got %d", acquired)
	}
}

func TestRateLimiterSetRate(t *testing.T) {
	rl := NewRateLimiter(10, 10)

	rl.SetRate(20)
	if rl.GetRate() != 20 {
		t.Errorf("SetRate(20) should set rate to 20, got %d", rl.GetRate())
	}

	rl.SetRate(0)
	if rl.IsEnabled() {
		t.Error("SetRate(0) should disable the rate limiter")
	}
}

func BenchmarkRateLimiterWait(b *testing.B) {
	rl := NewRateLimiter(0, 0) // Disabled for benchmark
	for i := 0; i < b.N; i++ {
		rl.Wait()
	}
}

func BenchmarkRateLimiterTryAcquire(b *testing.B) {
	rl := NewRateLimiter(1000000, 1000000) // High rate for benchmark
	for i := 0; i < b.N; i++ {
		rl.TryAcquire()
	}
}
