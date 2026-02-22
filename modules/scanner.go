// scanner.go - Chrome 135 Anti-Detection with uTLS Fingerprint
//
// This module provides:
// - Chrome 135 TLS fingerprint via uTLS (JA3/JA4 spoofing)
// - Chrome 135 browser headers in exact order
// - Smart jitter for natural request patterns
//
// Used by request.go to bypass Cloudflare and other WAFs.
// Updated: January 2026

package modules

import (
	"context"
	"encoding/base64"
	"fmt"
	"ipmap/config"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/proxy"
)

// ====================================================================
// CHROME 135 USER-AGENT POOL (Windows/macOS/Linux - Jan 2026)
// ====================================================================

var chrome135UserAgents = []string{
	// Windows 11 24H2 - Chrome 135 (latest stable)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.6998.88 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.6998.117 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.6998.178 Safari/537.36",

	// macOS 15 Sequoia - Chrome 135
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.6998.88 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.6998.117 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.6998.178 Safari/537.36",

	// Linux - Chrome 135
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.6998.88 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.6998.117 Safari/537.36",

	// Chrome 134 variants (fallback - one version behind)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",

	// Chrome 133 variants (fallback - two versions behind)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
}

// Chrome 135 sec-ch-ua variants (includes GREASE tokens)
var chrome135SecChUA = []string{
	`"Google Chrome";v="135", "Chromium";v="135", "Not-A.Brand";v="8"`,
	`"Chromium";v="135", "Google Chrome";v="135", "Not-A.Brand";v="8"`,
	`"Not-A.Brand";v="8", "Chromium";v="135", "Google Chrome";v="135"`,
	`"Google Chrome";v="135", "Not-A.Brand";v="8", "Chromium";v="135"`,
	// Chrome 134 fallback
	`"Google Chrome";v="134", "Chromium";v="134", "Not-A.Brand";v="8"`,
	`"Chromium";v="134", "Google Chrome";v="134", "Not-A.Brand";v="8"`,
}

// Platform values for sec-ch-ua-platform
var chromePlatforms = []string{
	`"Windows"`,
	`"macOS"`,
	`"Linux"`,
}

// Accept-Language variants (natural distribution)
var acceptLanguages = []string{
	"en-US,en;q=0.9",
	"en-GB,en;q=0.9",
	"en-US,en;q=0.9,tr;q=0.8",
	"en-US,en;q=0.9,de;q=0.8",
	"en-US,en;q=0.9,fr;q=0.8",
	"en-US,en;q=0.9,es;q=0.8",
	"en-US,en;q=0.9,ja;q=0.8",
	"en;q=0.9",
}

// Referer sources for more realistic requests
var refererSources = []string{
	"https://www.google.com/",
	"https://www.bing.com/",
	"https://duckduckgo.com/",
	"https://search.yahoo.com/",
	"",
}

// ====================================================================
// CHROME 135 HEADERS
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

// NewRandomChromeProfile creates a random Chrome 135 profile
func NewRandomChromeProfile() *ChromeHeaderProfile {
	ua := config.GetRandomString(chrome135UserAgents)
	platform := config.GetRandomString(chromePlatforms)

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
		SecChUA:         config.GetRandomString(chrome135SecChUA),
		SecChUAMobile:   "?0",
		SecChUAPlatform: platform,
		AcceptLanguage:  config.GetRandomString(acceptLanguages),
		Referer:         config.GetRandomString(refererSources),
	}
}

// AddRealChromeHeaders adds Chrome 135 headers in the exact real browser order
// Header order is checked by Cloudflare and other WAFs
func AddRealChromeHeaders(req *http.Request, profile *ChromeHeaderProfile) {
	if profile == nil {
		profile = NewRandomChromeProfile()
	}

	// Chrome 135's REAL header order (captured from DevTools)
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

	// 8. Accept-Encoding (includes zstd - critical for Chrome 135!)
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
// UTLS TRANSPORT (Chrome 135 TLS Fingerprint)
// ====================================================================

// UTLSTransport wraps utls for Chrome 135 TLS fingerprint
type UTLSTransport struct {
	proxyURL *url.URL
	timeout  time.Duration
}

// NewUTLSTransport creates a new transport with Chrome 135 fingerprint
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

	// Note: HTTP/2 transport is not used because we force HTTP/1.1 via ALPN
	// This avoids protocol mismatch issues with Go's http.Transport

	return t, nil
}

// DialTLSContext creates a TLS connection with Chrome 135 fingerprint
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

	// uTLS handshake with Chrome 135 fingerprint
	// Use custom spec to force HTTP/1.1 via ALPN (avoid HTTP/2 issues with Go's http.Transport)
	tlsConn := utls.UClient(conn, &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: config.InsecureSkipVerify,
	}, utls.HelloCustom)

	// Apply Chrome fingerprint spec
	spec, err := utls.UTLSIdToSpec(utls.HelloChrome_Auto)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get Chrome spec: %w", err)
	}

	// Modify ALPN extension to force HTTP/1.1 only
	for i, ext := range spec.Extensions {
		if alpn, ok := ext.(*utls.ALPNExtension); ok {
			alpn.AlpnProtocols = []string{"http/1.1"}
			spec.Extensions[i] = alpn
			break
		}
	}

	if err := tlsConn.ApplyPreset(&spec); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to apply Chrome preset: %w", err)
	}

	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	return tlsConn, nil
}

func (t *UTLSTransport) dialViaProxy(ctx context.Context, network, addr string) (net.Conn, error) {
	proxyAddr := t.proxyURL.Host
	dialer := &net.Dialer{Timeout: t.timeout}

	// SOCKS5 proxy support
	if t.proxyURL.Scheme == "socks5" {
		var auth *proxy.Auth
		if t.proxyURL.User != nil {
			password, _ := t.proxyURL.User.Password()
			auth = &proxy.Auth{
				User:     t.proxyURL.User.Username(),
				Password: password,
			}
		}

		socks5Dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, dialer)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}

		return socks5Dialer.Dial(network, addr)
	}

	// HTTP/HTTPS proxy (CONNECT method)
	conn, err := dialer.DialContext(ctx, "tcp", proxyAddr)
	if err != nil {
		return nil, err
	}

	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)

	// Add proxy authentication if provided
	if t.proxyURL.User != nil {
		password, _ := t.proxyURL.User.Password()
		auth := t.proxyURL.User.Username() + ":" + password
		encoded := base64Encode(auth)
		connectReq += "Proxy-Authorization: Basic " + encoded + "\r\n"
	}

	connectReq += "\r\n"

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

	return conn, nil
}

// base64Encode encodes a string to base64
func base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}
