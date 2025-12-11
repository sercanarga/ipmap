package tools

import (
	"ipmap/config"
	"ipmap/modules"
	"testing"
)

func TestFindIPValidation(t *testing.T) {
	// Save original config
	originalWorkers := config.Workers
	defer func() { config.Workers = originalWorkers }()

	tests := []struct {
		name     string
		ipBlocks []string
		workers  int
		wantErr  bool
	}{
		{
			name:     "Valid single CIDR",
			ipBlocks: []string{"192.168.1.0/30"},
			workers:  10,
			wantErr:  false,
		},
		{
			name:     "Valid multiple CIDRs",
			ipBlocks: []string{"192.168.1.0/30", "10.0.0.0/30"},
			workers:  10,
			wantErr:  false,
		},
		{
			name:     "Invalid CIDR",
			ipBlocks: []string{"invalid"},
			workers:  10,
			wantErr:  true, // Should log error but not panic
		},
		{
			name:     "Empty blocks",
			ipBlocks: []string{},
			workers:  10,
			wantErr:  true,
		},
		{
			name:     "Mixed valid and invalid",
			ipBlocks: []string{"192.168.1.0/30", "invalid", "10.0.0.0/30"},
			workers:  10,
			wantErr:  false, // Valid blocks should still work
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Workers = tt.workers

			// Calculate expected IP count
			var expectedIPs int
			for _, block := range tt.ipBlocks {
				ips, err := modules.CalcIPAddress(block)
				if err == nil {
					expectedIPs += len(ips)
				}
			}

			// Note: We can't easily test FindIP directly since it calls ResolveSite
			// which makes network requests. Instead we test the IP calculation logic.
			if tt.wantErr && expectedIPs > 0 {
				t.Errorf("Expected no valid IPs for error case, got %d", expectedIPs)
			}
		})
	}
}

func TestFindIPWorkerCalculation(t *testing.T) {
	tests := []struct {
		name           string
		ipCount        int
		workers        int
		timeout        int
		expectedMinSec int
	}{
		{
			name:           "Small scan",
			ipCount:        50,
			workers:        100,
			timeout:        2000,
			expectedMinSec: 1, // Minimum is 1
		},
		{
			name:           "Large scan",
			ipCount:        1000,
			workers:        100,
			timeout:        2000,
			expectedMinSec: 20,
		},
		{
			name:           "High workers",
			ipCount:        500,
			workers:        500,
			timeout:        2000,
			expectedMinSec: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workerCount := tt.workers
			if workerCount <= 0 {
				workerCount = 100
			}
			estimatedSeconds := (tt.ipCount / workerCount) * tt.timeout / 1000
			if estimatedSeconds < 1 {
				estimatedSeconds = 1
			}

			if estimatedSeconds < tt.expectedMinSec {
				t.Errorf("Estimated %d seconds, want at least %d", estimatedSeconds, tt.expectedMinSec)
			}
		})
	}
}

func TestFindIPInterruptData(t *testing.T) {
	interruptData := modules.NewInterruptData()

	// Test that interrupt data can be set
	testBlocks := []string{"192.168.1.0/24", "10.0.0.0/24"}
	interruptData.IPBlocks = testBlocks

	if len(interruptData.IPBlocks) != 2 {
		t.Errorf("Expected 2 IP blocks, got %d", len(interruptData.IPBlocks))
	}

	// Test cancellation
	if interruptData.IsCancelled() {
		t.Error("Should not be cancelled initially")
	}

	interruptData.Cancel()

	if !interruptData.IsCancelled() {
		t.Error("Should be cancelled after Cancel()")
	}
}
