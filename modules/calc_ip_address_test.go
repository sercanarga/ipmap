package modules

import (
	"testing"
)

func TestCalcIPAddress(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		wantCount   int
		wantError   bool
		description string
	}{
		{
			name:        "Valid /24 network",
			cidr:        "192.168.1.0/24",
			wantCount:   254, // Excluding network and broadcast
			wantError:   false,
			description: "Should return 254 usable IPs",
		},
		{
			name:        "Valid /30 network",
			cidr:        "10.0.0.0/30",
			wantCount:   2, // Only 2 usable IPs
			wantError:   false,
			description: "Should return 2 usable IPs",
		},
		{
			name:        "Single IP /32",
			cidr:        "8.8.8.8/32",
			wantCount:   1,
			wantError:   false,
			description: "Should handle single IP without panic",
		},
		{
			name:        "Valid /28 network",
			cidr:        "172.16.0.0/28",
			wantCount:   14,
			wantError:   false,
			description: "Should return 14 usable IPs",
		},
		{
			name:        "Invalid CIDR format",
			cidr:        "invalid",
			wantCount:   0,
			wantError:   true,
			description: "Should return error for invalid CIDR",
		},
		{
			name:        "Invalid IP address",
			cidr:        "999.999.999.999/24",
			wantCount:   0,
			wantError:   true,
			description: "Should return error for invalid IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := CalcIPAddress(tt.cidr)

			if tt.wantError {
				if err == nil {
					t.Errorf("CalcIPAddress() expected error but got none for %s", tt.cidr)
				}
				return
			}

			if err != nil {
				t.Errorf("CalcIPAddress() unexpected error: %v", err)
				return
			}

			if len(ips) != tt.wantCount {
				t.Errorf("CalcIPAddress() got %d IPs, want %d for %s (%s)",
					len(ips), tt.wantCount, tt.cidr, tt.description)
			}

			// Verify IPs are valid
			if len(ips) > 0 {
				if ips[0] == "" {
					t.Errorf("CalcIPAddress() returned empty IP string")
				}
			}
		})
	}
}

func TestCalcIPAddressRange(t *testing.T) {
	// Test that IPs are in correct range
	ips, err := CalcIPAddress("192.168.1.0/30")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"192.168.1.1", "192.168.1.2"}
	if len(ips) != len(expected) {
		t.Fatalf("Expected %d IPs, got %d", len(expected), len(ips))
	}

	for i, ip := range ips {
		if ip != expected[i] {
			t.Errorf("Expected IP %s at index %d, got %s", expected[i], i, ip)
		}
	}
}

func BenchmarkCalcIPAddress(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = CalcIPAddress("192.168.1.0/24")
	}
}
