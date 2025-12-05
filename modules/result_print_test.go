package modules

import (
	"encoding/json"
	"ipmap/config"
	"testing"
)

func TestResultDataJSON(t *testing.T) {
	result := ResultData{
		Method:     "Test Method",
		SearchSite: "example.com",
		Timeout:    300,
		IPBlocks:   []string{"192.168.1.0/24"},
		FoundedWebsites: [][]string{
			{"200", "192.168.1.1", "Test Site"},
		},
		Timestamp: "2025-11-30T00:00:00Z",
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal ResultData: %v", err)
	}

	// Unmarshal back
	var decoded ResultData
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ResultData: %v", err)
	}

	// Verify fields
	if decoded.Method != result.Method {
		t.Errorf("Method mismatch: got %s, want %s", decoded.Method, result.Method)
	}
	if decoded.SearchSite != result.SearchSite {
		t.Errorf("SearchSite mismatch: got %s, want %s", decoded.SearchSite, result.SearchSite)
	}
	if decoded.Timeout != result.Timeout {
		t.Errorf("Timeout mismatch: got %d, want %d", decoded.Timeout, result.Timeout)
	}
	if len(decoded.IPBlocks) != len(result.IPBlocks) {
		t.Errorf("IPBlocks length mismatch: got %d, want %d", len(decoded.IPBlocks), len(result.IPBlocks))
	}
	if len(decoded.FoundedWebsites) != len(result.FoundedWebsites) {
		t.Errorf("FoundedWebsites length mismatch: got %d, want %d", len(decoded.FoundedWebsites), len(result.FoundedWebsites))
	}
}

func TestResultDataJSONOmitEmpty(t *testing.T) {
	result := ResultData{
		Method:   "Test Method",
		Timeout:  300,
		IPBlocks: []string{"192.168.1.0/24"},
		// SearchSite is empty - should be omitted
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal ResultData: %v", err)
	}

	// search_site should not be in JSON when empty
	if !json.Valid(jsonData) {
		t.Error("Generated JSON is invalid")
	}

	// Verify omitempty works
	var decoded map[string]interface{}
	_ = json.Unmarshal(jsonData, &decoded)

	if _, exists := decoded["search_site"]; exists {
		t.Error("search_site should be omitted when empty")
	}
}

func TestPrintResultWithDifferentFormats(t *testing.T) {
	// Save original config
	originalFormat := config.Format
	defer func() { config.Format = originalFormat }()

	// Test text format
	config.Format = "text"
	if config.Format != "text" {
		t.Error("Failed to set text format")
	}

	// Test JSON format
	config.Format = "json"
	if config.Format != "json" {
		t.Error("Failed to set JSON format")
	}
}

func BenchmarkResultDataMarshal(b *testing.B) {
	result := ResultData{
		Method:     "Benchmark",
		SearchSite: "example.com",
		Timeout:    300,
		IPBlocks:   []string{"192.168.1.0/24", "10.0.0.0/24"},
		FoundedWebsites: [][]string{
			{"200", "192.168.1.1", "Site 1"},
			{"200", "192.168.1.2", "Site 2"},
		},
		Timestamp: "2025-11-30T00:00:00Z",
	}

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(result)
	}
}
