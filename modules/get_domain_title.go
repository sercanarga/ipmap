package modules

import (
	"ipmap/config"
)

// GetDomainTitle fetches and extracts the title from a domain's HTML page.
// It tries HTTPS first, then HTTP, and finally with www prefix.
// Returns [title, responseTime] or empty slice on failure.
func GetDomainTitle(url string) []string {
	// Try HTTPS first with longer timeout (slow CDNs)
	config.InfoLog("Resolving domain: %s", url)
	config.VerboseLog("Trying HTTPS for domain: %s", url)
	getTitle := RequestFunc("https://"+url, url, config.DefaultDomainTimeout)

	// If HTTPS fails, try HTTP
	if len(getTitle) == 0 {
		config.VerboseLog("HTTPS failed, trying HTTP for domain: %s", url)
		getTitle = RequestFunc("http://"+url, url, config.DefaultDomainTimeout)
	}

	// If still no response, try with www prefix
	if len(getTitle) == 0 {
		config.VerboseLog("Trying with www prefix: www.%s", url)
		getTitle = RequestFunc("https://www."+url, url, config.DefaultDomainTimeout)
		if len(getTitle) == 0 {
			getTitle = RequestFunc("http://www."+url, url, config.DefaultDomainTimeout)
		}
	}

	// If still no response, return empty
	if len(getTitle) == 0 {
		config.ErrorLog("Failed to resolve domain: %s", url)
		config.ErrorLog("Possible causes:")
		config.ErrorLog("  1. Domain is down or not responding")
		config.ErrorLog("  2. Firewall/proxy blocking the connection")
		config.ErrorLog("  3. Network connectivity issues")
		config.ErrorLog("  4. Domain requires authentication")
		config.ErrorLog("\nTry running with -v flag for detailed logs")
		return []string{}
	}

	config.VerboseLog("Response received: Status=%s, Time=%sms", getTitle[0], getTitle[3])

	title := ExtractTitle(getTitle[2])
	if title != "" {
		config.VerboseLog("Title found: %s", title)
		return []string{title, getTitle[3]}
	}

	// If no title found but we got a response, use domain name as title
	// This allows the scan to continue even if title extraction fails (e.g., 403 errors)
	config.VerboseLog("No <title> tag found, using domain as title")
	return []string{url, getTitle[3]}
}
