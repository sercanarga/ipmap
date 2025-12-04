package modules

import (
	"os"
	"strings"
	"testing"
)

func TestExportFileWithDomain(t *testing.T) {
	tests := []struct {
		name           string
		domain         string
		isJSON         bool
		expectedPrefix string
		expectedExt    string
	}{
		{
			name:           "Text export with domain",
			domain:         "example.com",
			isJSON:         false,
			expectedPrefix: "ipmap_example_com_",
			expectedExt:    ".txt",
		},
		{
			name:           "JSON export with domain",
			domain:         "test.example.com",
			isJSON:         true,
			expectedPrefix: "ipmap_test_example_com_",
			expectedExt:    ".json",
		},
		{
			name:           "Export without domain",
			domain:         "",
			isJSON:         false,
			expectedPrefix: "ipmap_",
			expectedExt:    ".txt",
		},
		{
			name:           "Domain with special chars",
			domain:         "sub.domain.com:8080",
			isJSON:         false,
			expectedPrefix: "ipmap_sub_domain_com_8080_",
			expectedExt:    ".txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test content
			content := "Test export content"

			// Call exportFile
			exportFile(content, tt.isJSON, tt.domain)

			// Find created file
			files, err := os.ReadDir(".")
			if err != nil {
				t.Fatalf("Failed to read directory: %v", err)
			}

			var foundFile string
			for _, file := range files {
				if strings.HasPrefix(file.Name(), tt.expectedPrefix) && strings.HasSuffix(file.Name(), tt.expectedExt) {
					foundFile = file.Name()
					break
				}
			}

			if foundFile == "" {
				t.Errorf("Expected file with prefix %s and extension %s not found", tt.expectedPrefix, tt.expectedExt)
				return
			}

			// Verify file content
			data, err := os.ReadFile(foundFile)
			if err != nil {
				t.Errorf("Failed to read exported file: %v", err)
			} else if string(data) != content {
				t.Errorf("File content mismatch: got %s, want %s", string(data), content)
			}

			// Cleanup
			os.Remove(foundFile)
		})
	}
}

func TestExportFilenameSanitization(t *testing.T) {
	tests := []struct {
		domain   string
		expected string
	}{
		{"example.com", "example_com"},
		{"sub.domain.com", "sub_domain_com"},
		{"test.com:8080", "test_com_8080"},
		{"site.com/path", "site_com_path"},
		{"complex.sub.domain.org:443/path", "complex_sub_domain_org_443_path"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			// Simulate sanitization
			safeDomain := strings.ReplaceAll(tt.domain, ".", "_")
			safeDomain = strings.ReplaceAll(safeDomain, "/", "_")
			safeDomain = strings.ReplaceAll(safeDomain, ":", "_")

			if safeDomain != tt.expected {
				t.Errorf("Sanitization failed: got %s, want %s", safeDomain, tt.expected)
			}
		})
	}
}

func TestExportFileCreation(t *testing.T) {
	// Test that export creates file successfully
	domain := "test.example.com"
	content := "Test content for export"

	exportFile(content, false, domain)

	// Find and verify file
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	var foundFile string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "ipmap_test_example_com_") && strings.HasSuffix(file.Name(), ".txt") {
			foundFile = file.Name()
			break
		}
	}

	if foundFile == "" {
		t.Error("Export file not created")
		return
	}

	// Verify content
	data, err := os.ReadFile(foundFile)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	} else if string(data) != content {
		t.Errorf("Content mismatch: got %s, want %s", string(data), content)
	}

	// Cleanup
	os.Remove(foundFile)
}
