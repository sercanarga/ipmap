package modules

import (
	"ipmap/config"
	"regexp"
	"strings"
)

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
			config.VerboseLog("Site found on %s: %s (Status: %s)", ip, title[1], explodeHttpCode[0])

			// Perform reverse DNS lookup
			hostname := ReverseDNS(ip)
			if hostname != "" {
				// Return with hostname: [status, ip, title, hostname]
				return []string{explodeHttpCode[0], requestSite[1], title[1], hostname}
			}

			return []string{explodeHttpCode[0], requestSite[1], title[1]}
		}
	}

	return []string{}
}
