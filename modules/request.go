package modules

import (
	"context"
	"crypto/tls"
	"io"
	"ipmap/config"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/corpix/uarand"
)

// HTTP client with lazy initialization
var (
	httpClient     *http.Client
	httpClientOnce sync.Once
	lastProxyURL   string
	lastDNSServers string
	clientMu       sync.RWMutex
)

// GetHTTPClient returns the HTTP client, creating or recreating if config changed
func GetHTTPClient() *http.Client {
	clientMu.RLock()
	currentProxy := config.ProxyURL
	currentDNS := strings.Join(config.DNSServers, ",")
	needsRecreate := httpClient != nil && (lastProxyURL != currentProxy || lastDNSServers != currentDNS)
	clientMu.RUnlock()

	if needsRecreate {
		clientMu.Lock()
		// Double-check after acquiring write lock
		if lastProxyURL != currentProxy || lastDNSServers != currentDNS {
			httpClient = createHTTPClientWithConfig()
			lastProxyURL = currentProxy
			lastDNSServers = currentDNS
			config.VerboseLog("HTTP client recreated with new config (Proxy: %s, DNS: %s)", currentProxy, currentDNS)
		}
		clientMu.Unlock()
		return httpClient
	}

	httpClientOnce.Do(func() {
		clientMu.Lock()
		defer clientMu.Unlock()
		httpClient = createHTTPClientWithConfig()
		lastProxyURL = config.ProxyURL
		lastDNSServers = strings.Join(config.DNSServers, ",")
	})

	return httpClient
}

// createCustomDialer creates a dialer with optional custom DNS servers
func createCustomDialer() *net.Dialer {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return dialer
}

// createDialContext creates a DialContext function with optional custom DNS
func createDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := createCustomDialer()

	if len(config.DNSServers) == 0 {
		return dialer.DialContext
	}

	// Custom DNS resolver
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			// Use first available custom DNS server
			for _, dns := range config.DNSServers {
				dnsAddr := strings.TrimSpace(dns)
				if !strings.Contains(dnsAddr, ":") {
					dnsAddr = dnsAddr + ":53"
				}
				conn, err := d.DialContext(ctx, "udp", dnsAddr)
				if err == nil {
					return conn, nil
				}
			}
			// Fallback to default
			return d.DialContext(ctx, network, address)
		},
	}

	config.VerboseLog("Using custom DNS servers: %v", config.DNSServers)

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// Split host and port
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return dialer.DialContext(ctx, network, addr)
		}

		// Resolve using custom DNS
		ips, err := resolver.LookupIPAddr(ctx, host)
		if err != nil || len(ips) == 0 {
			// Fallback to normal resolution
			return dialer.DialContext(ctx, network, addr)
		}

		// Try each resolved IP
		for _, ip := range ips {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
		}

		// Fallback
		return dialer.DialContext(ctx, network, addr)
	}
}

func createHTTPClientWithConfig() *http.Client {
	// Calculate connection pool size based on worker count
	maxConns := config.Workers
	if maxConns < 100 {
		maxConns = 100
	}
	maxConnsPerHost := maxConns / 10
	if maxConnsPerHost < 10 {
		maxConnsPerHost = 10
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		MaxIdleConns:          maxConns,
		MaxIdleConnsPerHost:   maxConnsPerHost,
		MaxConnsPerHost:       maxConnsPerHost * 2, // Allow more active connections
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext:           createDialContext(),
		ForceAttemptHTTP2:     true,
		DisableKeepAlives:     false, // Keep connections alive for reuse
	}

	// Configure proxy if specified
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err != nil {
			config.ErrorLog("Invalid proxy URL '%s': %v", config.ProxyURL, err)
		} else {
			transport.Proxy = http.ProxyURL(proxyURL)
			config.VerboseLog("Using proxy: %s", config.ProxyURL)
		}
	}

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			for key, val := range via[0].Header {
				if _, ok := req.Header[key]; !ok {
					req.Header[key] = val
				}
			}
			return nil
		},
	}
}

func RequestFunc(ip string, url string, timeout int) []string {
	return RequestFuncWithRetry(ip, url, timeout, config.MaxRetries)
}

func RequestFuncWithRetry(ip string, url string, timeout int, maxRetries int) []string {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			config.VerboseLog("Retry attempt %d/%d for %s", attempt, maxRetries, ip)
			// Exponential backoff
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
		}

		n := time.Now()

		req, err := http.NewRequest("GET", ip, nil)
		if err != nil {
			lastErr = err
			config.VerboseLog("Failed to create request: %v", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
		req = req.WithContext(ctx)

		// Set Host header for virtual hosting
		if url != "" {
			req.Host = url
		}

		// Set realistic browser headers to avoid bot detection
		ua := uarand.GetRandom()
		req.Header.Set("User-Agent", ua)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")

		// Randomize Accept-Language to avoid fingerprinting
		languages := []string{
			"en-US,en;q=0.9",
			"en-GB,en;q=0.9",
			"en-US,en;q=0.9,tr;q=0.8",
			"de-DE,de;q=0.9,en;q=0.8",
			"fr-FR,fr;q=0.9,en;q=0.8",
		}
		req.Header.Set("Accept-Language", languages[time.Now().UnixNano()%int64(len(languages))])

		req.Header.Set("Accept-Encoding", "identity") // No compression to avoid decompression issues
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Upgrade-Insecure-Requests", "1")
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "none")
		req.Header.Set("Sec-Fetch-User", "?1")
		req.Header.Set("Cache-Control", "max-age=0")

		// Randomize browser version fingerprint
		chromeVersions := []string{
			`"Chromium";v="120", "Not_A Brand";v="24"`,
			`"Chromium";v="119", "Not_A Brand";v="24"`,
			`"Chromium";v="121", "Not_A Brand";v="24"`,
			`"Google Chrome";v="120", "Chromium";v="120"`,
		}
		req.Header.Set("Sec-Ch-Ua", chromeVersions[time.Now().UnixNano()%int64(len(chromeVersions))])
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")

		// Randomize platform
		platforms := []string{`"Windows"`, `"macOS"`, `"Linux"`}
		req.Header.Set("Sec-Ch-Ua-Platform", platforms[time.Now().UnixNano()%int64(len(platforms))])

		resp, err := GetHTTPClient().Do(req)

		if err != nil {
			cancel() // Cancel on error
			lastErr = err
			config.VerboseLog("Request error (attempt %d): %v", attempt+1, err)
			continue
		}

		// Read body with limit to prevent memory issues
		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 1MB limit
		resp.Body.Close()
		cancel() // Cancel context after body is read

		if err != nil {
			lastErr = err
			config.VerboseLog("Failed to read response body: %v", err)
			continue
		}

		// Build response string similar to httputil.DumpResponse
		var responseBuilder strings.Builder
		responseBuilder.WriteString(resp.Proto)
		responseBuilder.WriteString(" ")
		responseBuilder.WriteString(resp.Status)
		responseBuilder.WriteString("\r\n")
		for key, values := range resp.Header {
			for _, value := range values {
				responseBuilder.WriteString(key)
				responseBuilder.WriteString(": ")
				responseBuilder.WriteString(value)
				responseBuilder.WriteString("\r\n")
			}
		}
		responseBuilder.WriteString("\r\n")
		responseBuilder.Write(bodyBytes)

		// Success! Return even for non-2xx status codes (let caller decide)
		elapsed := time.Since(n).Milliseconds()
		if attempt > 0 {
			config.VerboseLog("Request succeeded on retry %d for %s", attempt, ip)
		}
		config.VerboseLog("Response: Status=%s, Size=%d bytes, Time=%dms", resp.Status, len(bodyBytes), elapsed)

		return []string{resp.Status, ip, responseBuilder.String(), strconv.FormatInt(elapsed, 10)}
	}

	// All retries failed
	if lastErr != nil {
		config.VerboseLog("Connection failed for %s: %v", url, lastErr)
	}
	return []string{}
}
