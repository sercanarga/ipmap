package modules

import (
	"context"
	"ipmap/config"
	"net"
	"time"
)

// ReverseDNS performs reverse DNS lookup for an IP address
func ReverseDNS(ip string) string {
	config.VerboseLog("Performing reverse DNS lookup for: %s", ip)

	// Set timeout for DNS lookup
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 2,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	names, err := resolver.LookupAddr(ctx, ip)
	if err != nil {
		config.VerboseLog("Reverse DNS lookup failed for %s: %v", ip, err)
		return ""
	}

	if len(names) > 0 {
		config.VerboseLog("Reverse DNS found for %s: %s", ip, names[0])
		return names[0]
	}

	return ""
}
