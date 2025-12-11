// scanner.go - Chrome 131 Anti-Detection with uTLS Fingerprint
//
// This module provides:
// - Chrome 131 TLS fingerprint via uTLS (JA3/JA4 spoofing)
// - Chrome 131 browser headers in exact order
// - Smart jitter for natural request patterns
//
// Used by request.go to bypass Cloudflare and other WAFs.

package modules

import (
	"context"
	"fmt"
	"ipmap/config"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// ====================================================================
// CHROME 131 USER-AGENT POOL (Windows/macOS/Linux - Dec 2025)
// ====================================================================

var chrome131UserAgents = []string{
	// Windows 10/11 - Chrome 131
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.108 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.139 Safari/537.36",

	// macOS - Chrome 131
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.108 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.139 Safari/537.36",

	// Linux - Chrome 131
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.108 Safari/537.36",

	// Chrome 130 variants (fallback)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
}

// Chrome 131 sec-ch-ua variants (includes GREASE)
var chrome131SecChUA = []string{
	`"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
	`"Chromium";v="131", "Google Chrome";v="131", "Not_A Brand";v="24"`,
	`"Not_A Brand";v="24", "Chromium";v="131", "Google Chrome";v="131"`,
	`"Google Chrome";v="131", "Not_A Brand";v="24", "Chromium";v="131"`,
}

// Platform values
var chrome131Platforms = []string{
	`"Windows"`,
	`"macOS"`,
	`"Linux"`,
}

// Accept-Language variants
var acceptLanguages = []string{
	"en-US,en;q=0.9",
	"en-GB,en;q=0.9",
	"en-US,en;q=0.9,tr;q=0.8",
	"en-US,en;q=0.9,de;q=0.8",
	"en-US,en;q=0.9,fr;q=0.8",
	"en;q=0.9",
}

// Referer sources for more realistic requests
var refererSources = []string{
	"https://www.google.com/",
	"https://www.bing.com/",
	"https://duckduckgo.com/",
	"",
}

// ====================================================================
// CHROME 131 HEADERS
// ====================================================================

// ChromeHeaderProfile holds a complete Chrome header profile
type ChromeHeaderProfile struct {
	UserAgent       string
	SecChUA         string
	SecChUAMobile   string
	SecChUAPlatform string
	AcceptLanguage  string
	Referer         string
}

// NewRandomChromeProfile creates a random Chrome 131 profile
func NewRandomChromeProfile() *ChromeHeaderProfile {
	ua := config.GetRandomString(chrome131UserAgents)
	platform := config.GetRandomString(chrome131Platforms)

	// Match platform with User-Agent
	if strings.Contains(ua, "Windows") {
		platform = `"Windows"`
	} else if strings.Contains(ua, "Macintosh") {
		platform = `"macOS"`
	} else if strings.Contains(ua, "Linux") {
		platform = `"Linux"`
	}

	return &ChromeHeaderProfile{
		UserAgent:       ua,
		SecChUA:         config.GetRandomString(chrome131SecChUA),
		SecChUAMobile:   "?0",
		SecChUAPlatform: platform,
		AcceptLanguage:  config.GetRandomString(acceptLanguages),
		Referer:         config.GetRandomString(refererSources),
	}
}

// AddRealChromeHeaders adds Chrome 131 headers in the exact real browser order
// Header order is checked by Cloudflare and other WAFs
func AddRealChromeHeaders(req *http.Request, profile *ChromeHeaderProfile) {
	if profile == nil {
		profile = NewRandomChromeProfile()
	}

	// Chrome 131's REAL header order (captured from DevTools)
	// Order is critical! Must match Chrome's actual order, not alphabetical

	// 1. Host (auto-added)

	// 2. Connection
	req.Header.Set("Connection", "keep-alive")

	// 3. sec-ch-ua series (in this order!)
	req.Header.Set("sec-ch-ua", profile.SecChUA)
	req.Header.Set("sec-ch-ua-mobile", profile.SecChUAMobile)
	req.Header.Set("sec-ch-ua-platform", profile.SecChUAPlatform)

	// 4. Upgrade-Insecure-Requests
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// 5. User-Agent
	req.Header.Set("User-Agent", profile.UserAgent)

	// 6. Accept
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")

	// 7. Sec-Fetch series (in this order!)
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")

	// 8. Accept-Encoding (includes zstd - critical for Chrome 131!)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")

	// 9. Accept-Language
	req.Header.Set("Accept-Language", profile.AcceptLanguage)

	// 10. Referer (adds legitimacy to request)
	if profile.Referer != "" {
		req.Header.Set("Referer", profile.Referer)
	}

	// 11. Cache-Control
	req.Header.Set("Cache-Control", "max-age=0")
}

// ====================================================================
// UTLS TRANSPORT (Chrome 131 TLS Fingerprint)
// ====================================================================

// UTLSTransport wraps utls for Chrome 131 TLS fingerprint
type UTLSTransport struct {
	proxyURL    *url.URL
	timeout     time.Duration
	h2Transport *http2.Transport
}

// NewUTLSTransport creates a new transport with Chrome 131 fingerprint
func NewUTLSTransport(proxyURL string, timeout time.Duration) (*UTLSTransport, error) {
	t := &UTLSTransport{
		timeout: timeout,
	}

	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		t.proxyURL = parsed
	}

	// Setup HTTP/2 transport
	t.h2Transport = &http2.Transport{
		ReadIdleTimeout: 30 * time.Second,
		PingTimeout:     15 * time.Second,
	}

	return t, nil
}

// DialTLSContext creates a TLS connection with Chrome 131 fingerprint
func (t *UTLSTransport) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	// Create TCP connection
	dialer := &net.Dialer{Timeout: t.timeout}
	var conn net.Conn

	if t.proxyURL != nil {
		conn, err = t.dialViaProxy(ctx, network, addr)
	} else {
		conn, err = dialer.DialContext(ctx, network, addr)
	}
	if err != nil {
		return nil, err
	}

	// uTLS handshake with Chrome 131 fingerprint
	tlsConn := utls.UClient(conn, &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	}, utls.HelloChrome_Auto) // Auto-selects latest Chrome fingerprint

	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	return tlsConn, nil
}

func (t *UTLSTransport) dialViaProxy(ctx context.Context, network, addr string) (net.Conn, error) {
	proxyAddr := t.proxyURL.Host
	dialer := &net.Dialer{Timeout: t.timeout}

	conn, err := dialer.DialContext(ctx, "tcp", proxyAddr)
	if err != nil {
		return nil, err
	}

	// HTTP CONNECT for HTTPS proxy
	if t.proxyURL.Scheme == "http" || t.proxyURL.Scheme == "https" {
		connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", addr, addr)
		if _, err := conn.Write([]byte(connectReq)); err != nil {
			conn.Close()
			return nil, err
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			return nil, err
		}
		if !strings.Contains(string(buf[:n]), "200") {
			conn.Close()
			return nil, fmt.Errorf("proxy CONNECT failed: %s", string(buf[:n]))
		}
	}

	return conn, nil
}

// GetTransport returns an http.Transport with uTLS dial function
func (t *UTLSTransport) GetTransport() *http.Transport {
	return &http.Transport{
		DialTLSContext:        t.DialTLSContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ForceAttemptHTTP2:     true,
	}
}
