package modules

import (
	"fmt"
	"ipmap/config"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

// ShuffleIPs randomizes the order of IP addresses to avoid sequential scanning patterns
func ShuffleIPs(ips []string) []string {
	return config.ShuffleStrings(ips)
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

	// Early exit flag and once guard for PrintResult
	var stopped bool
	var printOnce sync.Once

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

			// Add random jitter to avoid detection patterns
			config.AddJitter()

			// Skip individual DNS lookup - will do batch DNS at the end
			site := GetSiteWithOptions(ip, domain, timeout, true)

			mu.Lock()
			scannedCount++
			mu.Unlock()

			if len(site) > 0 {
				// Check if cancelled before printing to avoid mixed output
				if interruptData != nil && interruptData.IsCancelled() {
					// Still add to websites for export even if cancelled
					interruptData.AddWebsite(site)
					return
				}

				// Format site info nicely for terminal output
				var siteInfo string
				if len(site) >= 4 {
					siteInfo = fmt.Sprintf("[%s] %s - %s [%s]", site[0], site[1], site[2], site[3])
				} else if len(site) >= 3 {
					siteInfo = fmt.Sprintf("[%s] %s - %s", site[0], site[1], site[2])
				} else {
					siteInfo = fmt.Sprintf("%v", site)
				}
				fmt.Printf("\n  [+] Found: %s\n", siteInfo)

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
					return
				}
			}

			// Only update progress bar if not cancelled
			if interruptData == nil || !interruptData.IsCancelled() {
				mu.Lock()
				_ = bar.Add(1)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	_ = bar.Finish()

	// Batch DNS lookup for all found sites (much faster than individual lookups)
	if len(foundSites) > 0 && !stopped && (interruptData == nil || !interruptData.IsCancelled()) {
		fmt.Printf("\n[*] Performing batch DNS lookup for %d found IPs...\n", len(foundSites))

		// Collect IPs that need DNS lookup (extract IP from URL like https://1.2.3.4)
		ipsToLookup := make([]string, 0, len(foundSites))
		for _, site := range foundSites {
			if len(site) >= 2 {
				ip := site[1]
				// Remove protocol prefix if present
				ip = strings.TrimPrefix(ip, "https://")
				ip = strings.TrimPrefix(ip, "http://")
				ipsToLookup = append(ipsToLookup, ip)
			}
		}

		// Perform batch DNS lookup
		dnsResults := BatchReverseDNS(ipsToLookup, DefaultDNSConcurrency)

		// Update foundSites with hostnames
		for i, site := range foundSites {
			if len(site) >= 2 {
				// Strip protocol prefix to match dnsResults key format
				lookupIP := strings.TrimPrefix(strings.TrimPrefix(site[1], "https://"), "http://")
				if hostname, ok := dnsResults[lookupIP]; ok && hostname != "" {
					// Append hostname to result
					if len(site) == 3 {
						foundSites[i] = append(site, hostname)
					}
				}
			}
		}

		// Also update interrupt data
		if interruptData != nil {
			interruptData.mu.Lock()
			for i, site := range interruptData.Websites {
				if len(site) >= 2 {
					lookupIP := strings.TrimPrefix(strings.TrimPrefix(site[1], "https://"), "http://")
					if hostname, ok := dnsResults[lookupIP]; ok && hostname != "" {
						if len(site) == 3 {
							interruptData.Websites[i] = append(site, hostname)
						}
					}
				}
			}
			interruptData.mu.Unlock()
		}

		fmt.Printf("[*] DNS lookup completed: %d/%d hostnames resolved\n", len(dnsResults), len(ipsToLookup))
	}

	// Print scan statistics
	fmt.Printf("\n[*] Scan Statistics: %d/%d IPs scanned, %d sites found\n", scannedCount, len(IPAddress), foundCount)

	// Process and print results (sync.Once guarantees single execution)
	printOnce.Do(func() {
		mu.Lock()
		wasEarlyStopped := stopped
		mu.Unlock()
		_ = bar.Finish()
		if wasEarlyStopped {
			PrintResult("Search Domain by ASN", DomainTitle, timeout, IPBlocks, foundSites, export)
		} else {
			PrintResult("Search All ASN/IP", DomainTitle, timeout, IPBlocks, foundSites, export)
		}
	})
}

// ResolveSiteWithCache performs scanning with cache support for resuming
func ResolveSiteWithCache(IPAddress []string, Websites [][]string, DomainTitle string, IPBlocks []string, domain string, con bool, export bool, timeout int, interruptData *InterruptData, cache *Cache) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	var foundSites [][]string
	var scannedCount, foundCount int

	// Add existing cached results
	if cache != nil {
		foundSites = append(foundSites, cache.GetResults()...)
		foundCount = len(foundSites)
	}

	shuffledIPs := ShuffleIPs(IPAddress)
	config.VerboseLog("IP addresses shuffled for firewall bypass")

	workerCount := config.Workers
	config.VerboseLog("Starting scan with %d concurrent workers", workerCount)
	sem := make(chan struct{}, workerCount)

	rateLimiter := NewRateLimiter(config.RateLimit, config.RateLimit*2)
	if rateLimiter.IsEnabled() {
		config.VerboseLog("Rate limiting enabled: %d requests/second", config.RateLimit)
	}

	bar := progressbar.NewOptions(len(shuffledIPs),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("[cyan][1/1][reset] Scanning IPs (Resuming)"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	var stopped bool
	var printOnce sync.Once
	saveCounter := 0

	for _, ip := range shuffledIPs {
		mu.Lock()
		if stopped {
			mu.Unlock()
			break
		}
		mu.Unlock()

		if interruptData != nil && interruptData.IsCancelled() {
			config.VerboseLog("Scan cancelled by user")
			break
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			if interruptData != nil && interruptData.IsCancelled() {
				return
			}
			mu.Lock()
			if stopped {
				mu.Unlock()
				return
			}
			mu.Unlock()

			rateLimiter.Wait()
			config.AddJitter()

			// Skip individual DNS lookup - will do batch DNS at the end
			site := GetSiteWithOptions(ip, domain, timeout, true)

			mu.Lock()
			scannedCount++

			// Save to cache
			if cache != nil {
				cache.AddScannedIP(ip)
				saveCounter++
				// Save cache every 50 IPs
				if saveCounter >= 50 {
					_ = cache.Save()
					saveCounter = 0
				}
			}
			mu.Unlock()

			if len(site) > 0 {
				if interruptData != nil && interruptData.IsCancelled() {
					interruptData.AddWebsite(site)
					if cache != nil {
						cache.AddResult(site)
					}
					return
				}

				var siteInfo string
				if len(site) >= 4 {
					siteInfo = fmt.Sprintf("[%s] %s - %s [%s]", site[0], site[1], site[2], site[3])
				} else if len(site) >= 3 {
					siteInfo = fmt.Sprintf("[%s] %s - %s", site[0], site[1], site[2])
				} else {
					siteInfo = fmt.Sprintf("%v", site)
				}
				fmt.Printf("\n  [+] Found: %s\n", siteInfo)

				mu.Lock()
				foundSites = append(foundSites, site)
				foundCount++
				if cache != nil {
					cache.AddResult(site)
				}
				mu.Unlock()

				if interruptData != nil {
					interruptData.AddWebsite(site)
				}

				if DomainTitle != "" && len(site) > 2 && site[2] == DomainTitle && !con {
					mu.Lock()
					stopped = true
					mu.Unlock()
					return
				}
			}

			if interruptData == nil || !interruptData.IsCancelled() {
				mu.Lock()
				_ = bar.Add(1)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	_ = bar.Finish()

	// Batch DNS lookup for all found sites (much faster than individual lookups)
	if len(foundSites) > 0 && !stopped && (interruptData == nil || !interruptData.IsCancelled()) {
		fmt.Printf("\n[*] Performing batch DNS lookup for %d found IPs...\n", len(foundSites))

		// Collect IPs that need DNS lookup (extract IP from URL like https://1.2.3.4)
		ipsToLookup := make([]string, 0, len(foundSites))
		for _, site := range foundSites {
			if len(site) >= 2 {
				ip := site[1]
				// Remove protocol prefix if present
				ip = strings.TrimPrefix(ip, "https://")
				ip = strings.TrimPrefix(ip, "http://")
				ipsToLookup = append(ipsToLookup, ip)
			}
		}

		// Perform batch DNS lookup
		dnsResults := BatchReverseDNS(ipsToLookup, DefaultDNSConcurrency)

		// Update foundSites with hostnames
		for i, site := range foundSites {
			if len(site) >= 2 {
				lookupIP := strings.TrimPrefix(strings.TrimPrefix(site[1], "https://"), "http://")
				if hostname, ok := dnsResults[lookupIP]; ok && hostname != "" {
					if len(site) == 3 {
						foundSites[i] = append(site, hostname)
					}
				}
			}
		}

		// Update interrupt data and cache
		if interruptData != nil {
			interruptData.mu.Lock()
			for i, site := range interruptData.Websites {
				if len(site) >= 2 {
					lookupIP := strings.TrimPrefix(strings.TrimPrefix(site[1], "https://"), "http://")
					if hostname, ok := dnsResults[lookupIP]; ok && hostname != "" {
						if len(site) == 3 {
							interruptData.Websites[i] = append(site, hostname)
						}
					}
				}
			}
			interruptData.mu.Unlock()
		}

		fmt.Printf("[*] DNS lookup completed: %d/%d hostnames resolved\n", len(dnsResults), len(ipsToLookup))
	}

	// Final save
	if cache != nil {
		cache.MarkCompleted()
		_ = cache.Save()
		config.InfoLog("Cache saved to: %s", cache.FilePath)
	}

	fmt.Printf("\n[*] Scan Statistics: %d/%d IPs scanned, %d sites found\n", scannedCount, len(IPAddress), foundCount)

	// Process and print results (sync.Once guarantees single execution)
	printOnce.Do(func() {
		mu.Lock()
		wasEarlyStopped := stopped
		mu.Unlock()
		_ = bar.Finish()
		if cache != nil {
			cache.MarkCompleted()
			_ = cache.Save()
		}
		if wasEarlyStopped {
			PrintResult("Search Domain by ASN (Resumed)", DomainTitle, timeout, IPBlocks, foundSites, export)
		} else {
			PrintResult("Search All ASN/IP (Resumed)", DomainTitle, timeout, IPBlocks, foundSites, export)
		}
	})
}
