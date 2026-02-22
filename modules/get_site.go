package modules

import (
	"ipmap/config"
	"strings"
)

// GetSite scans an IP address for a website and extracts its title.
// Returns [status, ip, title] or [status, ip, title, hostname] if reverse DNS succeeds.
// When skipDNS is true, DNS lookup is skipped (for batch DNS processing later).
func GetSite(ip string, domain string, timeout int) []string {
	return GetSiteWithOptions(ip, domain, timeout, false)
}

// GetSiteWithOptions scans an IP address with configurable options.
// skipDNS: if true, skips individual DNS lookup (use BatchReverseDNS later for better performance)
func GetSiteWithOptions(ip string, domain string, timeout int, skipDNS bool) []string {
	// Try HTTPS first (modern sites)
	config.VerboseLog("Scanning IP: %s (HTTPS)", ip)
	requestSite := RequestFunc("https://"+ip, domain, timeout)

	// If HTTPS fails, try HTTP
	if len(requestSite) == 0 {
		config.VerboseLog("HTTPS failed for %s, trying HTTP", ip)
		requestSite = RequestFunc("http://"+ip, domain, timeout)
	}

	if len(requestSite) > 0 {
		title := ExtractTitle(requestSite[2])
		if title != "" {
			explodeHttpCode := strings.Split(requestSite[0], " ")
			if len(explodeHttpCode) == 0 {
				config.VerboseLog("Malformed HTTP status for %s", ip)
				return []string{}
			}

			config.VerboseLog("Site found on %s: %s (Status: %s)", ip, title, explodeHttpCode[0])

			// Skip DNS lookup if requested (for batch processing)
			if skipDNS {
				return []string{explodeHttpCode[0], requestSite[1], title}
			}

			// Perform reverse DNS lookup
			hostname := ReverseDNS(ip)
			if hostname != "" {
				// Return with hostname: [status, ip, title, hostname]
				return []string{explodeHttpCode[0], requestSite[1], title, hostname}
			}

			return []string{explodeHttpCode[0], requestSite[1], title}
		}
	}

	return []string{}
}
