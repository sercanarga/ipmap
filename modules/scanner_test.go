package modules

import (
	"ipmap/config"
	"net/http"
	"testing"
	"time"
)

// ====================================================================
// SCANNER.GO TESTS
// ====================================================================

func TestNewRandomChromeProfile(t *testing.T) {
	profile := NewRandomChromeProfile()

	if profile == nil {
		t.Fatal("NewRandomChromeProfile returned nil")
	}

	// Check UserAgent is not empty
	if profile.UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}

	// Check sec-ch-ua is not empty
	if profile.SecChUA == "" {
		t.Error("SecChUA should not be empty")
	}

	// Check sec-ch-ua-mobile
	if profile.SecChUAMobile != "?0" {
		t.Errorf("SecChUAMobile should be '?0', got '%s'", profile.SecChUAMobile)
	}

	// Check platform is valid
	validPlatforms := map[string]bool{
		`"Windows"`: true,
		`"macOS"`:   true,
		`"Linux"`:   true,
	}
	if !validPlatforms[profile.SecChUAPlatform] {
		t.Errorf("Invalid SecChUAPlatform: %s", profile.SecChUAPlatform)
	}

	// Check AcceptLanguage is not empty
	if profile.AcceptLanguage == "" {
		t.Error("AcceptLanguage should not be empty")
	}
}

func TestAddRealChromeHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	profile := NewRandomChromeProfile()
	AddRealChromeHeaders(req, profile)

	// Check required headers
	requiredHeaders := []string{
		"Connection",
		"sec-ch-ua",
		"sec-ch-ua-mobile",
		"sec-ch-ua-platform",
		"Upgrade-Insecure-Requests",
		"User-Agent",
		"Accept",
		"Sec-Fetch-Site",
		"Sec-Fetch-Mode",
		"Sec-Fetch-User",
		"Sec-Fetch-Dest",
		"Accept-Encoding",
		"Accept-Language",
		"Cache-Control",
	}

	for _, header := range requiredHeaders {
		if req.Header.Get(header) == "" {
			t.Errorf("Header '%s' should be set", header)
		}
	}

	// Check Accept-Encoding contains zstd (Chrome 135 specific)
	acceptEncoding := req.Header.Get("Accept-Encoding")
	if acceptEncoding != "gzip, deflate, br, zstd" {
		t.Errorf("Accept-Encoding should include zstd, got: %s", acceptEncoding)
	}
}

func TestAddRealChromeHeadersWithNilProfile(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Should not panic with nil profile
	AddRealChromeHeaders(req, nil)

	// Headers should still be set with auto-generated profile
	if req.Header.Get("User-Agent") == "" {
		t.Error("User-Agent should be set even with nil profile")
	}
}

func TestConfigAddJitter(t *testing.T) {
	start := time.Now()

	// Call jitter from config
	config.AddJitter()

	elapsed := time.Since(start)

	// Should be at most MaxJitterMs + tolerance
	if elapsed > time.Duration(config.MaxJitterMs+100)*time.Millisecond {
		t.Errorf("Jitter should be at most ~%dms, was %v", config.MaxJitterMs, elapsed)
	}
}

func TestConfigGetRandomString(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}

	// Run multiple times to ensure no panic
	for i := 0; i < 100; i++ {
		result := config.GetRandomString(slice)
		found := false
		for _, s := range slice {
			if s == result {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetRandomString returned value not in slice: %s", result)
		}
	}

	// Test empty slice
	empty := config.GetRandomString([]string{})
	if empty != "" {
		t.Errorf("GetRandomString on empty slice should return empty string, got: %s", empty)
	}
}

func TestConfigGetRandomInt(t *testing.T) {
	max := 10

	// Run multiple times
	for i := 0; i < 100; i++ {
		result := config.GetRandomInt(max)
		if result < 0 || result >= max {
			t.Errorf("GetRandomInt(%d) returned %d, want 0 <= x < %d", max, result, max)
		}
	}

	// Test zero max
	zero := config.GetRandomInt(0)
	if zero != 0 {
		t.Errorf("GetRandomInt(0) should return 0, got: %d", zero)
	}
}

func TestConfigShuffleStrings(t *testing.T) {
	original := []string{"a", "b", "c", "d", "e"}
	shuffled := config.ShuffleStrings(original)

	// Check length is same
	if len(shuffled) != len(original) {
		t.Errorf("Shuffled length %d != original length %d", len(shuffled), len(original))
	}

	// Check all elements are present
	seen := make(map[string]bool)
	for _, s := range shuffled {
		seen[s] = true
	}
	for _, s := range original {
		if !seen[s] {
			t.Errorf("Element %s missing from shuffled result", s)
		}
	}
}

func TestChromeUserAgentVariety(t *testing.T) {
	seen := make(map[string]bool)

	// Generate multiple profiles
	for i := 0; i < 50; i++ {
		profile := NewRandomChromeProfile()
		seen[profile.UserAgent] = true
	}

	// Should have some variety
	if len(seen) < 3 {
		t.Errorf("Expected variety in User-Agents, only got %d unique", len(seen))
	}
}

func TestChromeSecChUAVariety(t *testing.T) {
	seen := make(map[string]bool)

	// Generate multiple profiles
	for i := 0; i < 50; i++ {
		profile := NewRandomChromeProfile()
		seen[profile.SecChUA] = true
	}

	// Should have some variety (we have 4 variants)
	if len(seen) < 2 {
		t.Errorf("Expected variety in sec-ch-ua, only got %d unique", len(seen))
	}
}

func BenchmarkNewRandomChromeProfile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRandomChromeProfile()
	}
}

func BenchmarkAddRealChromeHeaders(b *testing.B) {
	profile := NewRandomChromeProfile()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		AddRealChromeHeaders(req, profile)
	}
}
