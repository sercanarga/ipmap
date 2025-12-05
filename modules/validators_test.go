package modules

import (
	"testing"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid CIDR", "192.168.1.0/24", true},
		{"Valid IPv6", "::1", true},
		{"Invalid IP", "999.999.999.999", false},
		{"Invalid CIDR", "192.168.1.0/33", false},
		{"Empty string", "", false},
		{"Random string", "not-an-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateIP(tt.input)
			if got != tt.want {
				t.Errorf("ValidateIP(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateASN(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid ASN uppercase", "AS13335", true},
		{"Valid ASN lowercase", "as13335", true},
		{"Valid short ASN", "AS1", true},
		{"Invalid - no AS prefix", "13335", false},
		{"Invalid - letters after AS", "ASABC", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateASN(tt.input)
			if got != tt.want {
				t.Errorf("ValidateASN(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid domain", "example.com", true},
		{"Valid subdomain", "sub.example.com", true},
		{"Empty string (optional)", "", true},
		{"With protocol", "https://example.com", true},
		{"With path", "example.com/path", true},
		{"Single part", "localhost", false},
		{"Too long", string(make([]byte, 255)) + ".com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateDomain(tt.input)
			if got != tt.want {
				t.Errorf("ValidateDomain(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateCIDRList(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{"Single CIDR", "192.168.1.0/24", 1, false},
		{"Multiple CIDRs", "192.168.1.0/24,10.0.0.0/8", 2, false},
		{"With spaces", "192.168.1.0/24, 10.0.0.0/8", 2, false},
		{"Invalid CIDR", "invalid", 0, true},
		{"Empty string", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateCIDRList(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCIDRList(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("ValidateCIDRList(%q) count = %d, want %d", tt.input, len(got), tt.wantCount)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Normal name", "test", "test"},
		{"With dots", "example.com", "example_com"},
		{"With slashes", "path/to/file", "path_to_file"},
		{"With colons", "file:name", "file_name"},
		{"Complex", "sub.domain.com:8080/path", "sub_domain_com_8080_path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateWorkerCount(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  int
	}{
		{"Normal value", 100, 100},
		{"Zero", 0, 1},
		{"Negative", -1, 1},
		{"Too high", 2000, 1000},
		{"Max allowed", 1000, 1000},
		{"Min allowed", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateWorkerCount(tt.input)
			if got != tt.want {
				t.Errorf("ValidateWorkerCount(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
