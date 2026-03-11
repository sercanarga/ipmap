package modules

import (
	"fmt"
	"net"
)

// MaxIPsPerBlock is the maximum number of IPs to generate from a CIDR block
// This prevents memory exhaustion from very large blocks (e.g., /8 networks)
const MaxIPsPerBlock = 1000000

// CalcIPAddress generates all usable IP addresses from a CIDR block.
// It excludes network and broadcast addresses for blocks larger than /30.
// Returns an error if the CIDR is invalid or would generate more than MaxIPsPerBlock IPs.
func CalcIPAddress(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	// Check for excessively large CIDR blocks to prevent memory exhaustion
	ones, bits := ipnet.Mask.Size()
	if bits-ones >= 20 { // 1 million or more IPs
		maxIPs := 1 << uint(bits-ones)
		return nil, fmt.Errorf("CIDR block /%d too large: would generate %d IPs (max: %d)", ones, maxIPs, MaxIPsPerBlock)
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// Handle single IP case (/32 or /128)
	if len(ips) <= 2 {
		return ips, nil
	}

	// Remove network and broadcast addresses
	return ips[1 : len(ips)-1], nil
}

// inc increments an IP address by one (in-place modification).
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
