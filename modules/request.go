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
)

var (
	// HTTP client with lazy initialization
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
			// Close idle connections before recreating to prevent leaks
			if httpClient != nil {
				httpClient.CloseIdleConnections()
			}
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

		// Pre-warm connection pool if proxy is configured
		if config.ProxyURL != "" {
			go preWarmConnectionPool(httpClient, config.ProxyURL)
		}
	})

	return httpClient
}

// preWarmConnectionPool establishes initial connections to the proxy
// This reduces latency for the first requests
func preWarmConnectionPool(client *http.Client, proxyURL string) {
	if client == nil || proxyURL == "" {
		return
	}

	config.VerboseLog("Pre-warming connection pool for proxy: %s", proxyURL)

	// Parse proxy to get host
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return
	}

	// Establish a few initial connections
	for i := 0; i < 3; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Simple HEAD request to establish connection
			req, err := http.NewRequestWithContext(ctx, "HEAD", "https://"+parsed.Host, nil)
			if err != nil {
				return
			}

			resp, err := client.Do(req)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}()
	}
}

// createCustomDialer creates a dialer with optimal connection settings
func createCustomDialer() *net.Dialer {
	dialer := &net.Dialer{
		Timeout:   time.Duration(config.DialTimeout) * time.Second,
		KeepAlive: 30 * time.Second,
		// DualStack enables both IPv4 and IPv6
		DualStack: config.EnableIPv6,
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
			d := net.Dialer{Timeout: 3 * time.Second}
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
	// Try to create uTLS client for Chrome 135 TLS fingerprint
	utlsTransport, err := NewUTLSTransport(config.ProxyURL, time.Duration(config.DialTimeout)*time.Second)
	if err != nil {
		config.VerboseLog("Failed to create uTLS transport: %v, using fallback", err)
		return createFallbackHTTPClient()
	}

	// Calculate optimized connection pool size based on worker count
	// More aggressive pooling for better performance
	maxConns := config.Workers * 2
	if maxConns < 200 {
		maxConns = 200
	}
	if maxConns > 1000 {
		maxConns = 1000
	}

	// Per-host connections (important for proxy mode)
	maxConnsPerHost := config.Workers
	if maxConnsPerHost < 50 {
		maxConnsPerHost = 50
	}
	if maxConnsPerHost > 200 {
		maxConnsPerHost = 200
	}

	// Create transport with uTLS dial function and optimized pooling
	transport := &http.Transport{
		DialTLSContext: utlsTransport.DialTLSContext,
		DialContext:    createDialContext(),
		// Connection Pool Settings
		MaxIdleConns:        maxConns,
		MaxIdleConnsPerHost: maxConnsPerHost,
		MaxConnsPerHost:     maxConnsPerHost * 2,
		// Longer idle timeout for better reuse
		IdleConnTimeout: 90 * time.Second,
		// Timeouts
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// HTTP/2 disabled for uTLS compatibility
		ForceAttemptHTTP2: false,
		// Keep connections alive for reuse
		DisableKeepAlives: false,
		// Enable compression for better performance
		DisableCompression: false,
		// Write buffer for better throughput
		WriteBufferSize: 64 * 1024, // 64KB
		ReadBufferSize:  64 * 1024, // 64KB
	}

	config.VerboseLog("Connection pool: MaxIdle=%d, MaxPerHost=%d, IdleTimeout=90s", maxConns, maxConnsPerHost)
	config.VerboseLog("Using uTLS transport with Chrome 135 TLS fingerprint")

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			// Preserve headers on redirect
			for key, val := range via[0].Header {
				if _, ok := req.Header[key]; !ok {
					req.Header[key] = val
				}
			}
			return nil
		},
	}
}

// createFallbackHTTPClient creates a standard HTTP client as fallback
func createFallbackHTTPClient() *http.Client {
	// Calculate optimized connection pool size (same as main client)
	maxConns := config.Workers * 2
	if maxConns < 200 {
		maxConns = 200
	}
	if maxConns > 1000 {
		maxConns = 1000
	}

	maxConnsPerHost := config.Workers
	if maxConnsPerHost < 50 {
		maxConnsPerHost = 50
	}
	if maxConnsPerHost > 200 {
		maxConnsPerHost = 200
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
			MinVersion:         tls.VersionTLS12,
			// TLS Session Cache for connection reuse
			ClientSessionCache: tls.NewLRUClientSessionCache(256),
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		// Connection Pool Settings
		MaxIdleConns:        maxConns,
		MaxIdleConnsPerHost: maxConnsPerHost,
		MaxConnsPerHost:     maxConnsPerHost * 2,
		// Longer idle timeout for better reuse
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext:           createDialContext(),
		ForceAttemptHTTP2:     true,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		// Buffer sizes for better throughput
		WriteBufferSize: 64 * 1024,
		ReadBufferSize:  64 * 1024,
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

	config.VerboseLog("Fallback client - Connection pool: MaxIdle=%d, MaxPerHost=%d", maxConns, maxConnsPerHost)

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
			// Backoff before retry
			time.Sleep(time.Duration(attempt*200) * time.Millisecond)
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

		// Use Chrome 135 headers from scanner.go for better anti-detection
		profile := NewRandomChromeProfile()
		AddRealChromeHeaders(req, profile)

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
