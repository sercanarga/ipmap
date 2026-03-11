package modules

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"ipmap/config"
)

// TestGenerateCacheFileName tests cache filename generation
func TestGenerateCacheFileName(t *testing.T) {
	tests := []struct {
		name    string
		asn     string
		domain  string
		wantASN bool
		wantDom bool
	}{
		{"With ASN", "AS13335", "", true, false},
		{"With domain", "", "example.com", false, true},
		{"With neither", "", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCacheFileName(tt.asn, tt.domain)

			// Should start with ipmap_
			if !strings.HasPrefix(result, "ipmap_") {
				t.Errorf("Filename should start with 'ipmap_', got: %s", result)
			}

			// Should end with _cache.json
			if !strings.HasSuffix(result, "_cache.json") {
				t.Errorf("Filename should end with '_cache.json', got: %s", result)
			}

			// Should contain ASN if provided
			if tt.wantASN && !strings.Contains(result, tt.asn) {
				t.Errorf("Filename should contain ASN '%s', got: %s", tt.asn, result)
			}

			// Should contain sanitized domain if provided
			if tt.wantDom && !strings.Contains(result, "example_com") {
				t.Errorf("Filename should contain sanitized domain, got: %s", result)
			}

			// When no ASN or domain, should have timestamp (numeric)
			if !tt.wantASN && !tt.wantDom {
				// Extract the part between ipmap_ and _cache.json
				middle := strings.TrimPrefix(result, "ipmap_")
				middle = strings.TrimSuffix(middle, "_cache.json")
				_, err := strconv.ParseInt(middle, 10, 64)
				if err != nil {
					t.Errorf("Filename timestamp should be numeric, got: %s (error: %v)", middle, err)
				}
			}
		})
	}
}

// TestOutputDirConfig tests output directory configuration
func TestOutputDirConfig(t *testing.T) {
	// Save original
	original := config.OutputDir
	defer func() { config.OutputDir = original }()

	// Test setting output dir
	config.OutputDir = "/tmp/exports"
	if config.OutputDir != "/tmp/exports" {
		t.Errorf("OutputDir not set correctly, got: %s", config.OutputDir)
	}

	// Test empty (default)
	config.OutputDir = ""
	if config.OutputDir != "" {
		t.Errorf("OutputDir should be empty, got: %s", config.OutputDir)
	}
}

// TestInsecureSkipVerifyConfig tests TLS verification configuration
func TestInsecureSkipVerifyConfig(t *testing.T) {
	// Save original
	original := config.InsecureSkipVerify
	defer func() { config.InsecureSkipVerify = original }()

	// Default should be true (backward compatibility)
	if !config.InsecureSkipVerify {
		t.Error("InsecureSkipVerify default should be true")
	}

	// Test setting to false
	config.InsecureSkipVerify = false
	if config.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be false when set")
	}
}

// TestCacheOperations tests cache CRUD operations
func TestCacheOperations(t *testing.T) {
	tmpFile := "test_cache_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".json"
	defer os.Remove(tmpFile)

	// Create new cache
	cache := NewCache(tmpFile)
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	// Set metadata
	cache.SetMetadata("AS13335", "example.com", "Example Domain", 2000, []string{"1.0.0.0/24"})

	// Add scanned IPs
	cache.AddScannedIP("1.0.0.1")
	cache.AddScannedIP("1.0.0.2")

	// Add results
	cache.AddResult([]string{"200", "https://1.0.0.1", "Example Site"})

	// Check IsScanned
	if !cache.IsScanned("1.0.0.1") {
		t.Error("IP 1.0.0.1 should be marked as scanned")
	}
	if cache.IsScanned("1.0.0.99") {
		t.Error("IP 1.0.0.99 should not be marked as scanned")
	}

	// Save cache
	if err := cache.Save(); err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}

	// Load cache
	loaded, err := LoadCache(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	// Verify data
	if loaded.Data.ASN != "AS13335" {
		t.Errorf("ASN mismatch: got %s, want AS13335", loaded.Data.ASN)
	}
	if len(loaded.Data.ScannedIPs) != 2 {
		t.Errorf("ScannedIPs count mismatch: got %d, want 2", len(loaded.Data.ScannedIPs))
	}
	if len(loaded.Data.Results) != 1 {
		t.Errorf("Results count mismatch: got %d, want 1", len(loaded.Data.Results))
	}

	// Test GetUnscannedIPs
	allIPs := []string{"1.0.0.1", "1.0.0.2", "1.0.0.3", "1.0.0.4"}
	unscanned := loaded.GetUnscannedIPs(allIPs)
	if len(unscanned) != 2 {
		t.Errorf("Unscanned IPs count mismatch: got %d, want 2", len(unscanned))
	}

	// Test MarkCompleted
	loaded.MarkCompleted()
	if !loaded.Data.Completed {
		t.Error("Cache should be marked as completed")
	}
}

// TestGetProgress tests accurate CIDR-based progress calculation
func TestGetProgress(t *testing.T) {
	tests := []struct {
		name          string
		ipBlocks      []string
		scannedIPs    []string
		results       [][]string
		wantScanned   int
		wantTotal     int
		wantResults   int
	}{
		{
			name:        "/24 block has 254 usable IPs",
			ipBlocks:    []string{"1.0.0.0/24"},
			scannedIPs:  []string{"1.0.0.1", "1.0.0.2"},
			results:     [][]string{{"200", "https://1.0.0.1", "Test"}},
			wantScanned: 2,
			wantTotal:   254,
			wantResults: 1,
		},
		{
			name:        "/30 block has 2 usable IPs",
			ipBlocks:    []string{"10.0.0.0/30"},
			scannedIPs:  []string{"10.0.0.1"},
			results:     nil,
			wantScanned: 1,
			wantTotal:   2,
			wantResults: 0,
		},
		{
			name:        "Multiple blocks sum correctly",
			ipBlocks:    []string{"1.0.0.0/24", "2.0.0.0/24"},
			scannedIPs:  []string{"1.0.0.1"},
			results:     nil,
			wantScanned: 1,
			wantTotal:   508, // 254 + 254
			wantResults: 0,
		},
		{
			name:        "Invalid block falls back to 256 estimate",
			ipBlocks:    []string{"invalid-cidr"},
			scannedIPs:  nil,
			results:     nil,
			wantScanned: 0,
			wantTotal:   256,
			wantResults: 0,
		},
		{
			name:        "No blocks returns scanned count as total",
			ipBlocks:    nil,
			scannedIPs:  []string{"1.0.0.1", "1.0.0.2", "1.0.0.3"},
			results:     nil,
			wantScanned: 3,
			wantTotal:   3,
			wantResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := "test_progress_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".json"
			defer os.Remove(tmpFile)

			cache := NewCache(tmpFile)
			if tt.ipBlocks != nil {
				cache.SetMetadata("", "", "", 2000, tt.ipBlocks)
			}
			for _, ip := range tt.scannedIPs {
				cache.AddScannedIP(ip)
			}
			for _, r := range tt.results {
				cache.AddResult(r)
			}

			scanned, total, results := cache.GetProgress()
			if scanned != tt.wantScanned {
				t.Errorf("scanned = %d, want %d", scanned, tt.wantScanned)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if results != tt.wantResults {
				t.Errorf("results = %d, want %d", results, tt.wantResults)
			}
		})
	}
}

// TestExtractTitle tests HTML title extraction
func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{"Simple title", "<html><head><title>Hello World</title></head></html>", "Hello World"},
		{"Title with entities", "<title>Test &amp; Demo</title>", "Test & Demo"},
		{"Title with whitespace", "<title>  Spaced Title  </title>", "Spaced Title"},
		{"No title", "<html><body>No title here</body></html>", ""},
		{"Empty title", "<title></title>", ""},
		{"Title with newlines", "<title>Multi\nLine\nTitle</title>", "Multi Line Title"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTitle(tt.html)
			if result != tt.expected {
				t.Errorf("ExtractTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}
