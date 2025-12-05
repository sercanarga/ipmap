package modules

import (
	"net"
	"regexp"
	"strings"
	"unicode"
)

// ValidateIP checks if the given string is a valid IP address or CIDR
func ValidateIP(ip string) bool {
	// Check if it's a CIDR notation
	if strings.Contains(ip, "/") {
		_, _, err := net.ParseCIDR(ip)
		return err == nil
	}

	// Check if it's a plain IP
	return net.ParseIP(ip) != nil
}

// ValidateASN checks if the given string is a valid ASN format
func ValidateASN(asn string) bool {
	// ASN format: AS followed by numbers (e.g., AS13335)
	matched, _ := regexp.MatchString(`^[Aa][Ss]\d{1,10}$`, asn)
	return matched
}

// ValidateDomain checks if the given string is a valid domain name
func ValidateDomain(domain string) bool {
	if domain == "" {
		return true // Empty domain is allowed (optional parameter)
	}

	// Remove protocol if present
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// Remove trailing slash and path
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port if present
	if idx := strings.LastIndex(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Basic domain validation
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// Check for valid characters and structure
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		// Each part must start and end with alphanumeric
		if !isAlphanumeric(rune(part[0])) || !isAlphanumeric(rune(part[len(part)-1])) {
			return false
		}
	}

	return true
}

// ValidateCIDRList validates a comma-separated list of CIDR notations
func ValidateCIDRList(cidrList string) ([]string, error) {
	blocks := strings.Split(cidrList, ",")
	var validBlocks []string

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		_, _, err := net.ParseCIDR(block)
		if err != nil {
			return nil, err
		}
		validBlocks = append(validBlocks, block)
	}

	return validBlocks, nil
}

// SanitizeFilename removes or replaces characters that are invalid in filenames
func SanitizeFilename(name string) string {
	// Characters not allowed in filenames on various systems
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}

	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Replace dots with underscores (except for extension)
	result = strings.ReplaceAll(result, ".", "_")

	// Remove any non-printable characters
	var sanitized strings.Builder
	for _, r := range result {
		if unicode.IsPrint(r) && r != ' ' {
			sanitized.WriteRune(r)
		}
	}

	return sanitized.String()
}

// ValidateWorkerCount validates and returns a safe worker count
func ValidateWorkerCount(workers int) int {
	if workers < 1 {
		return 1
	}
	if workers > 1000 {
		return 1000
	}
	return workers
}

// ValidateTimeout validates and returns a safe timeout value
func ValidateTimeout(timeout int) int {
	if timeout < 100 {
		return 100 // Minimum 100ms
	}
	if timeout > 60000 {
		return 60000 // Maximum 60 seconds
	}
	return timeout
}

// isAlphanumeric checks if a rune is alphanumeric
func isAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
