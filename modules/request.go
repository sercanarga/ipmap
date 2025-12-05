package modules

import (
	"context"
	"crypto/tls"
	"io"
	"ipmap/config"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/corpix/uarand"
)

// Reusable HTTP client with connection pooling
var httpClient *http.Client

func init() {
	httpClient = createHTTPClient()
}

func createHTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			// Allow more cipher suites for compatibility
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		// Connection pooling
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		// Timeouts
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// Custom dialer with timeout
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		// Enable HTTP/2
		ForceAttemptHTTP2: true,
	}

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
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,tr;q=0.8")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Upgrade-Insecure-Requests", "1")
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "none")
		req.Header.Set("Sec-Fetch-User", "?1")
		req.Header.Set("Cache-Control", "max-age=0")
		req.Header.Set("Sec-Ch-Ua", `"Chromium";v="120", "Not_A Brand";v="24"`)
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)

		resp, err := httpClient.Do(req)

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
