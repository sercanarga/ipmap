package modules

import (
	"context"
	"ipmap/config"
	"net"
	"strings"
	"sync"
	"time"
)

// DefaultDNSConcurrency is the default number of parallel DNS lookups
const DefaultDNSConcurrency = 20

// BatchReverseDNS performs parallel reverse DNS lookups for multiple IPs
// Returns a map of IP -> hostname for successful lookups
func BatchReverseDNS(ips []string, concurrency int) map[string]string {
	if len(ips) == 0 {
		return make(map[string]string)
	}

	if concurrency <= 0 {
		concurrency = DefaultDNSConcurrency
	}

	// Cap concurrency to avoid overwhelming DNS servers
	if concurrency > 50 {
		concurrency = 50
	}

	config.VerboseLog("Starting batch DNS lookup for %d IPs with concurrency %d", len(ips), concurrency)

	results := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore for concurrency control
	sem := make(chan struct{}, concurrency)

	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()

			// Acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			hostname := ReverseDNS(ip)

			if hostname != "" {
				mu.Lock()
				results[ip] = hostname
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	config.VerboseLog("Batch DNS lookup completed: %d/%d hostnames resolved", len(results), len(ips))

	return results
}

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
			// Use custom DNS servers if configured
			if len(config.DNSServers) > 0 {
				for _, dns := range config.DNSServers {
					dnsAddr := strings.TrimSpace(dns)
					if !strings.Contains(dnsAddr, ":") {
						dnsAddr = dnsAddr + ":53"
					}
					conn, err := d.DialContext(ctx, network, dnsAddr)
					if err == nil {
						return conn, nil
					}
				}
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
