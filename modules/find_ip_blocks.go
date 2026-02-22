package modules

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"ipmap/config"
	"net/http"
	"strings"
	"time"
)

// RIPEStatResponse represents the response from RIPE Stat API
type RIPEStatResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Data       struct {
		Prefixes []struct {
			Prefix string `json:"prefix"`
		} `json:"prefixes"`
	} `json:"data"`
}

// FindIPBlocks queries RIPE Stat API to find all IP blocks for a given ASN.
// Returns formatted string containing route entries (CIDR notation).
// Falls back to RADB if RIPE Stat fails.
func FindIPBlocks(asn string) string {
	// Normalize ASN format - RIPE API expects just the number
	asnNumber := strings.TrimPrefix(strings.ToUpper(asn), "AS")

	// Try RIPE Stat API first (more reliable, no Cloudflare protection)
	result := fetchRIPEPrefixes(asnNumber, asn)
	if result != "" {
		return result
	}

	config.VerboseLog("RIPE Stat API failed for ASN %s, trying RADB fallback", asn)

	// Fallback to RADB (may not work due to Cloudflare)
	radbURL := "https://www.radb.net/query?advanced_query=1&keywords=" + asn + "&-T+option=&ip_option=&-i=1&-i+option=origin"
	output := RequestFunc(radbURL, "www.radb.net", config.DefaultAPITimeout)
	if len(output) > 2 {
		return output[2]
	}

	return ""
}

// fetchRIPEPrefixes makes a direct request to RIPE Stat API with proper handling
func fetchRIPEPrefixes(asnNumber, asn string) string {
	ripeURL := "https://stat.ripe.net/data/announced-prefixes/data.json?resource=AS" + asnNumber

	config.VerboseLog("Fetching IP prefixes for ASN %s from RIPE Stat API...", asn)

	// Use longer timeout for RIPE API (20 seconds) to handle slow network connections
	// This prevents freezing by ensuring the request eventually times out
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ripeURL, nil)
	if err != nil {
		config.VerboseLog("Failed to create RIPE request: %v", err)
		return ""
	}

	// Set headers for RIPE API
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Create a dedicated HTTP client for ASN queries to avoid interference with scan client
	asnClient := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			IdleConnTimeout:       30 * time.Second,
			DisableKeepAlives:     true, // Don't reuse connections for one-off API calls
		},
	}

	resp, err := asnClient.Do(req)
	if err != nil {
		config.VerboseLog("RIPE API request error: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		config.VerboseLog("RIPE API returned status %d", resp.StatusCode)
		return ""
	}

	// Handle gzip encoding
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			config.VerboseLog("Failed to create gzip reader: %v", err)
			return ""
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Read and parse JSON
	bodyBytes, err := io.ReadAll(io.LimitReader(reader, 2*1024*1024)) // 2MB limit
	if err != nil {
		config.VerboseLog("Failed to read RIPE response: %v", err)
		return ""
	}

	var response RIPEStatResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		config.VerboseLog("Failed to parse RIPE JSON: %v", err)
		return ""
	}

	if response.Status != "ok" || len(response.Data.Prefixes) == 0 {
		config.VerboseLog("RIPE API returned no prefixes for ASN %s", asn)
		return ""
	}

	// Build formatted output similar to RADB format (for compatibility with existing regex)
	// Only include IPv4 prefixes - skip IPv6 (contains ":")
	var result strings.Builder
	ipv4Count := 0
	for _, prefix := range response.Data.Prefixes {
		// Skip IPv6 prefixes unless explicitly enabled
		if strings.Contains(prefix.Prefix, ":") && !config.EnableIPv6 {
			continue
		}
		result.WriteString("route:          ")
		result.WriteString(prefix.Prefix)
		result.WriteString("\n")
		ipv4Count++
	}

	if ipv4Count == 0 {
		config.VerboseLog("RIPE API returned no IPv4 prefixes for ASN %s", asn)
		return ""
	}

	config.VerboseLog("Found %d IPv4 prefixes for ASN %s via RIPE Stat API (skipped %d IPv6)", ipv4Count, asn, len(response.Data.Prefixes)-ipv4Count)
	return result.String()
}
