package config

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVerboseLog(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test with verbose enabled
	Verbose = true
	VerboseLog("Test message: %s", "hello")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[VERBOSE]") {
		t.Error("VerboseLog should print when Verbose is true")
	}
	if !strings.Contains(output, "Test message: hello") {
		t.Error("VerboseLog should print formatted message")
	}

	// Test with verbose disabled
	r, w, _ = os.Pipe()
	os.Stdout = w

	Verbose = false
	VerboseLog("Should not print")

	w.Close()
	os.Stdout = old

	buf.Reset()
	_, _ = io.Copy(&buf, r)
	output = buf.String()

	if output != "" {
		t.Error("VerboseLog should not print when Verbose is false")
	}
}

func TestErrorLog(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ErrorLog("Error: %d", 404)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[ERROR]") {
		t.Error("ErrorLog should print [ERROR] prefix")
	}
	if !strings.Contains(output, "Error: 404") {
		t.Error("ErrorLog should print formatted message")
	}
}

func TestInfoLog(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	InfoLog("Info: %s", "test")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Error("InfoLog should print [INFO] prefix")
	}
	if !strings.Contains(output, "Info: test") {
		t.Error("InfoLog should print formatted message")
	}
}

func TestConfigDefaults(t *testing.T) {
	if MaxRetries != 0 {
		t.Errorf("MaxRetries default should be 0 (retries disabled), got %d", MaxRetries)
	}
	if Workers != 100 {
		t.Errorf("Workers default should be 100, got %d", Workers)
	}
}

func TestWorkerConfiguration(t *testing.T) {
	// Save original
	original := Workers
	defer func() { Workers = original }()

	tests := []struct {
		name  string
		value int
	}{
		{"Minimum workers", 1},
		{"Default workers", 100},
		{"High workers", 500},
		{"Maximum workers", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Workers = tt.value
			if Workers != tt.value {
				t.Errorf("Workers = %d, want %d", Workers, tt.value)
			}
		})
	}
}

func TestWarnLog(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	WarnLog("Warning: %s %d", "test", 42)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[WARN]") {
		t.Error("WarnLog should print [WARN] prefix")
	}
	if !strings.Contains(output, "Warning: test 42") {
		t.Error("WarnLog should print formatted message")
	}
}

func TestLoadConfigFile(t *testing.T) {
	// Create temp YAML config
	tmpFile, err := os.CreateTemp("", "ipmap_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	yamlContent := `workers: 200
rate_limit: 50
proxy: "http://127.0.0.1:8080"
dns_servers:
  - "8.8.8.8"
  - "1.1.1.1"
ipv6: true
verbose: true
format: "json"
`
	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfigFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	if cfg.Workers != 200 {
		t.Errorf("Workers = %d, want 200", cfg.Workers)
	}
	if cfg.RateLimit != 50 {
		t.Errorf("RateLimit = %d, want 50", cfg.RateLimit)
	}
	if cfg.Proxy != "http://127.0.0.1:8080" {
		t.Errorf("Proxy = %s, want http://127.0.0.1:8080", cfg.Proxy)
	}
	if len(cfg.DNSServers) != 2 {
		t.Errorf("DNSServers count = %d, want 2", len(cfg.DNSServers))
	}
	if !cfg.IPv6 {
		t.Error("IPv6 should be true")
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %s, want json", cfg.Format)
	}
}

func TestLoadConfigFileInvalid(t *testing.T) {
	// Non-existent file
	_, err := LoadConfigFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadConfigFile should fail for non-existent file")
	}

	// Invalid YAML
	tmpFile, err := os.CreateTemp("", "ipmap_bad_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("invalid: yaml: content: [[[")
	tmpFile.Close()

	_, err = LoadConfigFile(tmpFile.Name())
	if err == nil {
		t.Error("LoadConfigFile should fail for invalid YAML")
	}
}

func TestApplyFileConfig(t *testing.T) {
	// Save originals
	origWorkers := Workers
	origRate := RateLimit
	origProxy := ProxyURL
	origDNS := DNSServers
	origIPv6 := EnableIPv6
	origVerbose := Verbose
	origFormat := Format
	defer func() {
		Workers = origWorkers
		RateLimit = origRate
		ProxyURL = origProxy
		DNSServers = origDNS
		EnableIPv6 = origIPv6
		Verbose = origVerbose
		Format = origFormat
	}()

	// Test nil config (should not panic)
	ApplyFileConfig(nil)

	// Test partial config (only workers)
	Workers = 100
	RateLimit = 0
	ApplyFileConfig(&FileConfig{Workers: 500})
	if Workers != 500 {
		t.Errorf("Workers = %d, want 500", Workers)
	}
	if RateLimit != 0 {
		t.Error("RateLimit should not change with zero value in config")
	}

	// Test full config
	ApplyFileConfig(&FileConfig{
		Workers:    300,
		RateLimit:  25,
		Proxy:      "socks5://localhost:1080",
		DNSServers: []string{"9.9.9.9"},
		IPv6:       true,
		Verbose:    true,
		Format:     "json",
	})
	if Workers != 300 {
		t.Errorf("Workers = %d, want 300", Workers)
	}
	if RateLimit != 25 {
		t.Errorf("RateLimit = %d, want 25", RateLimit)
	}
	if ProxyURL != "socks5://localhost:1080" {
		t.Errorf("ProxyURL = %s, want socks5://localhost:1080", ProxyURL)
	}
	if len(DNSServers) != 1 || DNSServers[0] != "9.9.9.9" {
		t.Errorf("DNSServers = %v, want [9.9.9.9]", DNSServers)
	}
	if !EnableIPv6 {
		t.Error("EnableIPv6 should be true")
	}
	if !Verbose {
		t.Error("Verbose should be true")
	}
	if Format != "json" {
		t.Errorf("Format = %s, want json", Format)
	}
}

func TestFindConfigFile(t *testing.T) {
	// Save current directory
	origDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "ipmap_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	defer os.Chdir(origDir)

	// Change to temp dir (no config files)
	os.Chdir(tmpDir)
	result := FindConfigFile()
	if result != "" {
		t.Errorf("FindConfigFile should return empty when no config exists, got: %s", result)
	}

	// Create config.yaml
	os.WriteFile("config.yaml", []byte("workers: 50\n"), 0644)
	result = FindConfigFile()
	if result != "config.yaml" {
		t.Errorf("FindConfigFile should find config.yaml, got: %s", result)
	}
}
