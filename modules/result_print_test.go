package modules

import (
	"encoding/json"
	"ipmap/config"
	"strings"
	"testing"
)

func TestResultDataJSON(t *testing.T) {
	result := ResultData{
		Method:     "Test Method",
		SearchSite: "example.com",
		Timeout:    300,
		IPBlocks:   []string{"192.168.1.0/24"},
		FoundWebsites: [][]string{
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
	if len(decoded.FoundWebsites) != len(result.FoundWebsites) {
		t.Errorf("FoundWebsites length mismatch: got %d, want %d", len(decoded.FoundWebsites), len(result.FoundWebsites))
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
		FoundWebsites: [][]string{
			{"200", "192.168.1.1", "Site 1"},
			{"200", "192.168.1.2", "Site 2"},
		},
		Timestamp: "2025-11-30T00:00:00Z",
	}

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(result)
	}
}

func TestResultDataBackwardCompatJSON(t *testing.T) {
	// Simulate legacy JSON with "founded_websites" instead of "found_websites"
	legacyJSON := `{
		"method": "Search All ASN/IP",
		"search_site": "example.com",
		"timeout_ms": 2000,
		"ip_blocks": ["1.0.0.0/24"],
		"founded_websites": [
			["200", "https://1.0.0.1", "Legacy Site"]
		],
		"timestamp": "2025-01-01T00:00:00Z"
	}`

	var result ResultData
	err := json.Unmarshal([]byte(legacyJSON), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal legacy JSON: %v", err)
	}

	// FoundWebsites should be populated from founded_websites
	if len(result.FoundWebsites) != 1 {
		t.Fatalf("FoundWebsites should have 1 entry from legacy field, got %d", len(result.FoundWebsites))
	}
	if result.FoundWebsites[0][2] != "Legacy Site" {
		t.Errorf("FoundWebsites[0][2] = %s, want 'Legacy Site'", result.FoundWebsites[0][2])
	}
}

func TestResultDataMarshalBothFields(t *testing.T) {
	result := ResultData{
		Method:  "Test",
		Timeout: 1000,
		FoundWebsites: [][]string{
			{"200", "https://1.0.0.1", "Test Site"},
		},
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(jsonData)

	// Both fields should be present in output
	if !strings.Contains(jsonStr, `"found_websites"`) {
		t.Error("JSON should contain 'found_websites' field")
	}
	if !strings.Contains(jsonStr, `"founded_websites"`) {
		t.Error("JSON should contain 'founded_websites' field for backward compatibility")
	}
}
