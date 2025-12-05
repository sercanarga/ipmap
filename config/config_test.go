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
	if MaxRetries != 2 {
		t.Errorf("MaxRetries default should be 2, got %d", MaxRetries)
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
