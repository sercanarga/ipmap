package modules

import (
	"fmt"
	"ipmap/config"
	"math/rand"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Package-level random generator (initialized once)
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
var rngMu sync.Mutex

// ShuffleIPs randomizes the order of IP addresses to avoid sequential scanning patterns
func ShuffleIPs(ips []string) []string {
	shuffled := make([]string, len(ips))
	copy(shuffled, ips)
	rngMu.Lock()
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	rngMu.Unlock()
	return shuffled
}

// AddJitter adds random delay to avoid detection patterns
func AddJitter(maxMs int) {
	if maxMs <= 0 {
		return
	}
	rngMu.Lock()
	jitter := rng.Intn(maxMs)
	rngMu.Unlock()
	time.Sleep(time.Duration(jitter) * time.Millisecond)
}

func ResolveSite(IPAddress []string, Websites [][]string, DomainTitle string, IPBlocks []string, domain string, con bool, export bool, timeout int, interruptData *InterruptData) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Use local slice to collect results (fix for slice passed by value issue)
	var foundSites [][]string

	// Scan statistics
	var scannedCount, foundCount int

	// Shuffle IPs to avoid sequential scanning patterns (firewall bypass)
	shuffledIPs := ShuffleIPs(IPAddress)
	config.VerboseLog("IP addresses shuffled for firewall bypass")

	// Use configurable worker pool size
	workerCount := config.Workers
	config.VerboseLog("Starting scan with %d concurrent workers", workerCount)
	sem := make(chan struct{}, workerCount)

	// Create rate limiter
	rateLimiter := NewRateLimiter(config.RateLimit, config.RateLimit*2)
	if rateLimiter.IsEnabled() {
		config.VerboseLog("Rate limiting enabled: %d requests/second", config.RateLimit)
	}

	// Create progress bar
	bar := progressbar.NewOptions(len(shuffledIPs),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("[cyan][1/1][reset] Scanning IPs"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// Atomic flag for early exit
	var stopped bool

	for _, ip := range shuffledIPs {
		// Check if already stopped or cancelled via Ctrl+C
		mu.Lock()
		if stopped {
			mu.Unlock()
			break
		}
		mu.Unlock()

		// Check for interrupt signal
		if interruptData != nil && interruptData.IsCancelled() {
			config.VerboseLog("Scan cancelled by user")
			break
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			// Check if stopped or cancelled before making request
			if interruptData != nil && interruptData.IsCancelled() {
				return
			}
			mu.Lock()
			if stopped {
				mu.Unlock()
				return
			}
			mu.Unlock()

			// Wait for rate limiter before making request
			rateLimiter.Wait()

			// Add random jitter (0-50ms) to avoid detection patterns
			AddJitter(50)

			site := GetSite(ip, domain, timeout)

			mu.Lock()
			scannedCount++
			mu.Unlock()

			if len(site) > 0 {

				fmt.Println("\n", site)
				mu.Lock()
				foundSites = append(foundSites, site)
				foundCount++
				mu.Unlock()

				// Add to interrupt data for Ctrl+C handling
				if interruptData != nil {
					interruptData.AddWebsite(site)
				}

				if DomainTitle != "" && len(site) > 2 && site[2] == DomainTitle && !con {
					mu.Lock()
					stopped = true
					mu.Unlock()
					_ = bar.Finish()
					PrintResult("Search Domain by ASN", DomainTitle, timeout, IPBlocks, foundSites, export)
					return
				}
			}

			mu.Lock()
			_ = bar.Add(1)
			mu.Unlock()
		}(ip)
	}

	wg.Wait()
	_ = bar.Finish()

	// Print scan statistics
	fmt.Printf("\n[*] Scan Statistics: %d/%d IPs scanned, %d sites found\n", scannedCount, len(IPAddress), foundCount)

	// Process and print results (only if not already printed)
	mu.Lock()
	if !stopped {
		mu.Unlock()
		PrintResult("Search All ASN/IP", DomainTitle, timeout, IPBlocks, foundSites, export)
	} else {
		mu.Unlock()
	}
}
