package modules

import (
	"context"
	"crypto/tls"
	"github.com/corpix/uarand"
	"ipmap/config"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

func RequestFunc(ip string, url string, timeout int) []string {
	return RequestFuncWithRetry(ip, url, timeout, config.MaxRetries)
}

func RequestFuncWithRetry(ip string, url string, timeout int, maxRetries int) []string {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			config.VerboseLog("Retry attempt %d/%d for %s", attempt, maxRetries, ip)
			// Longer exponential backoff
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

		req.Host = url

		// Set realistic browser headers to avoid bot detection
		req.Header.Set("User-Agent", uarand.GetRandom())
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Upgrade-Insecure-Requests", "1")
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "none")
		req.Header.Set("Cache-Control", "max-age=0")

		// Create client with TLS config and redirect support
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		}

		resp, err := client.Do(req)
		cancel() // Cancel context after request

		if err != nil {
			lastErr = err
			config.VerboseLog("Request error (attempt %d): %v", attempt+1, err)
			continue
		}
		defer resp.Body.Close()

		last, err := httputil.DumpResponse(resp, true)
		if err != nil {
			lastErr = err
			config.VerboseLog("Failed to dump response: %v", err)
			continue
		}

		// Success!
		if attempt > 0 {
			config.VerboseLog("Request succeeded on retry %d for %s", attempt, ip)
		}
		return []string{resp.Status, ip, string(last), strconv.FormatInt(time.Since(n).Milliseconds(), 10)}
	}

	// All retries failed - only show error in verbose mode
	if lastErr != nil {
		config.VerboseLog("Connection failed for %s: %v", url, lastErr)
	}
	return []string{}
}
