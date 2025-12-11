package modules

import (
	"html"
	"ipmap/config"
	"regexp"
	"strings"
)

// GetSite scans an IP address for a website and extracts its title.
// Returns [status, ip, title] or [status, ip, title, hostname] if reverse DNS succeeds.
func GetSite(ip string, domain string, timeout int) []string {
	// Try HTTPS first (modern sites)
	config.VerboseLog("Scanning IP: %s (HTTPS)", ip)
	requestSite := RequestFunc("https://"+ip, domain, timeout)

	// If HTTPS fails, try HTTP
	if len(requestSite) == 0 {
		config.VerboseLog("HTTPS failed for %s, trying HTTP", ip)
		requestSite = RequestFunc("http://"+ip, domain, timeout)
	}

	if len(requestSite) > 0 {
		re := regexp.MustCompile(`(?s).*?<title>(.*?)</title>.*`)
		title := re.FindStringSubmatch(requestSite[2])
		if len(title) > 0 {
			explodeHttpCode := strings.Split(requestSite[0], " ")
			if len(explodeHttpCode) == 0 {
				config.VerboseLog("Malformed HTTP status for %s", ip)
				return []string{}
			}

			// Decode HTML entities (e.g., &amp; -> &, &lt; -> <)
			decodedTitle := html.UnescapeString(title[1])
			config.VerboseLog("Site found on %s: %s (Status: %s)", ip, decodedTitle, explodeHttpCode[0])

			// Perform reverse DNS lookup
			hostname := ReverseDNS(ip)
			if hostname != "" {
				// Return with hostname: [status, ip, title, hostname]
				return []string{explodeHttpCode[0], requestSite[1], decodedTitle, hostname}
			}

			return []string{explodeHttpCode[0], requestSite[1], decodedTitle}
		}
	}

	return []string{}
}
