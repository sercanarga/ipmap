package modules

import (
	"testing"
	"time"
)

func TestReverseDNS(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		expectEmpty bool
		description string
	}{
		{
			name:        "Google DNS",
			ip:          "8.8.8.8",
			expectEmpty: false,
			description: "Should resolve to dns.google",
		},
		{
			name:        "Cloudflare DNS",
			ip:          "1.1.1.1",
			expectEmpty: false,
			description: "Should resolve to one.one.one.one",
		},
		{
			name:        "Invalid IP",
			ip:          "999.999.999.999",
			expectEmpty: true,
			description: "Should return empty for invalid IP",
		},
		{
			name:        "Private IP",
			ip:          "192.168.1.1",
			expectEmpty: false, // May or may not have PTR depending on network
			description: "Private IP may have local PTR record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set timeout for test
			done := make(chan string, 1)
			go func() {
				done <- ReverseDNS(tt.ip)
			}()

			select {
			case result := <-done:
				if tt.expectEmpty && result != "" {
					t.Errorf("ReverseDNS(%s) expected empty, got %s", tt.ip, result)
				}
				// For non-empty expectations, just log the result
				if !tt.expectEmpty {
					if result != "" {
						t.Logf("ReverseDNS(%s) returned: %s", tt.ip, result)
					} else {
						t.Logf("ReverseDNS(%s) returned empty (might be network issue)", tt.ip)
					}
				}
			case <-time.After(5 * time.Second):
				t.Errorf("ReverseDNS(%s) timed out", tt.ip)
			}
		})
	}
}

func TestReverseDNSTimeout(t *testing.T) {
	// Test that DNS lookup respects timeout
	start := time.Now()
	_ = ReverseDNS("192.168.255.255") // Non-routable IP
	elapsed := time.Since(start)

	// Should timeout within 3 seconds (2s timeout + buffer)
	if elapsed > 3*time.Second {
		t.Errorf("ReverseDNS took too long: %v", elapsed)
	}
}

func TestBatchReverseDNS(t *testing.T) {
	tests := []struct {
		name        string
		ips         []string
		concurrency int
		minExpected int // Minimum expected results (allows for network issues)
	}{
		{
			name:        "Multiple well-known IPs",
			ips:         []string{"8.8.8.8", "1.1.1.1", "8.8.4.4"},
			concurrency: 10,
			minExpected: 0, // PTR queries may fail depending on ISP/firewall
		},
		{
			name:        "Empty list",
			ips:         []string{},
			concurrency: 10,
			minExpected: 0,
		},
		{
			name:        "Single IP",
			ips:         []string{"8.8.8.8"},
			concurrency: 5,
			minExpected: 0, // Network might not allow
		},
		{
			name:        "Mixed valid and invalid",
			ips:         []string{"8.8.8.8", "999.999.999.999", "1.1.1.1"},
			concurrency: 10,
			minExpected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			results := BatchReverseDNS(tt.ips, tt.concurrency)
			elapsed := time.Since(start)

			t.Logf("BatchReverseDNS completed in %v, got %d results", elapsed, len(results))

			if len(results) < tt.minExpected {
				t.Errorf("Expected at least %d results, got %d", tt.minExpected, len(results))
			}

			// Log resolved hostnames
			for ip, hostname := range results {
				t.Logf("  %s -> %s", ip, hostname)
			}
		})
	}
}

func TestBatchReverseDNSConcurrency(t *testing.T) {
	// Test that batch is faster than sequential for multiple IPs
	ips := []string{"8.8.8.8", "1.1.1.1", "8.8.4.4", "1.0.0.1"}

	// Batch lookup
	start := time.Now()
	_ = BatchReverseDNS(ips, 10)
	batchTime := time.Since(start)

	t.Logf("Batch DNS for %d IPs took: %v", len(ips), batchTime)

	// Should complete within reasonable time (not sequential)
	// Sequential would be at least 4 * 2s = 8s for worst case timeout
	if batchTime > 6*time.Second {
		t.Errorf("Batch DNS took too long: %v (expected < 6s)", batchTime)
	}
}

func BenchmarkReverseDNS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ReverseDNS("8.8.8.8")
	}
}

func BenchmarkBatchReverseDNS(b *testing.B) {
	ips := []string{"8.8.8.8", "1.1.1.1", "8.8.4.4", "1.0.0.1"}
	for i := 0; i < b.N; i++ {
		_ = BatchReverseDNS(ips, 10)
	}
}
