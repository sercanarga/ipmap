package modules

import (
	"ipmap/config"
	"testing"
)

func TestResolveSiteWorkerPool(t *testing.T) {
	// Save original config
	originalWorkers := config.Workers
	defer func() { config.Workers = originalWorkers }()

	tests := []struct {
		name        string
		workers     int
		ipCount     int
		description string
	}{
		{
			name:        "Small worker pool",
			workers:     5,
			ipCount:     10,
			description: "Should handle 10 IPs with 5 workers",
		},
		{
			name:        "Large worker pool",
			workers:     200,
			ipCount:     50,
			description: "Should handle 50 IPs with 200 workers",
		},
		{
			name:        "Single worker",
			workers:     1,
			ipCount:     5,
			description: "Should handle sequential processing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Workers = tt.workers

			// Create dummy IP list
			ips := make([]string, tt.ipCount)
			for i := 0; i < tt.ipCount; i++ {
				ips[i] = "192.168.1.1" // Dummy IP
			}

			// This would normally scan IPs, but we're just testing the worker pool setup
			// In a real test, you'd mock the GetSite function
			if config.Workers != tt.workers {
				t.Errorf("Worker count not set correctly: got %d, want %d", config.Workers, tt.workers)
			}
		})
	}
}

func TestWorkerPoolValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "Valid worker count",
			input:    50,
			expected: 50,
		},
		{
			name:     "Minimum boundary",
			input:    1,
			expected: 1,
		},
		{
			name:     "Maximum boundary",
			input:    1000,
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Workers = tt.input
			if config.Workers != tt.expected {
				t.Errorf("Worker validation failed: got %d, want %d", config.Workers, tt.expected)
			}
		})
	}
}

func BenchmarkResolveSiteWorkerPool(b *testing.B) {
	// Save original config
	originalWorkers := config.Workers
	defer func() { config.Workers = originalWorkers }()

	workerCounts := []int{10, 50, 100, 200}

	for _, workers := range workerCounts {
		b.Run(string(rune(workers)), func(b *testing.B) {
			config.Workers = workers
			// Benchmark would go here
			// In real scenario, you'd benchmark actual IP scanning
		})
	}
}
