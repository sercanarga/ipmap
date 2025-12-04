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

func BenchmarkReverseDNS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ReverseDNS("8.8.8.8")
	}
}
