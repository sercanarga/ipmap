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

// scanConfig holds all configuration for a scan operation
type scanConfig struct {
	ipAddress     []string
	domainTitle   string
	ipBlocks      []string
	domain        string
	con           bool
	export        bool
	timeout       int
	interruptData *InterruptData
	cache         *Cache
	methodPrefix  string // "Search" or "Search ... (Resumed)"
}

// runScan is the core scanning engine shared by ResolveSite and ResolveSiteWithCache
func runScan(cfg *scanConfig) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Use local slice to collect results
	var foundSites [][]string

	// Scan statistics
	var scannedCount, foundCount int

	// Add existing cached results if resuming
	if cfg.cache != nil {
		foundSites = append(foundSites, cfg.cache.GetResults()...)
		foundCount = len(foundSites)
	}

	// Shuffle IPs to avoid sequential scanning patterns (firewall bypass)
	shuffledIPs := ShuffleIPs(cfg.ipAddress)
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
	description := "[cyan][1/1][reset] Scanning IPs"
	if cfg.cache != nil {
		description = "[cyan][1/1][reset] Scanning IPs (Resuming)"
	}
	bar := progressbar.NewOptions(len(shuffledIPs),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription(description),
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
	saveCounter := 0

	for _, ip := range shuffledIPs {
		// Check if already stopped or cancelled via Ctrl+C
		mu.Lock()
		if stopped {
			mu.Unlock()
			break
		}
		mu.Unlock()

		// Check for interrupt signal
		if cfg.interruptData != nil && cfg.interruptData.IsCancelled() {
			config.VerboseLog("Scan cancelled by user")
			break
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			// Check if stopped or cancelled before making request
			if cfg.interruptData != nil && cfg.interruptData.IsCancelled() {
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
			site := GetSiteWithOptions(ip, cfg.domain, cfg.timeout, true)

			mu.Lock()
			scannedCount++

			// Save to cache periodically
			if cfg.cache != nil {
				cfg.cache.AddScannedIP(ip)
				saveCounter++
				// Save cache every 50 IPs
				if saveCounter >= 50 {
					_ = cfg.cache.Save()
					saveCounter = 0
				}
			}
			mu.Unlock()

			if len(site) > 0 {
				// Check if cancelled before printing to avoid mixed output
				if cfg.interruptData != nil && cfg.interruptData.IsCancelled() {
					// Still add to websites for export even if cancelled
					cfg.interruptData.AddWebsite(site)
					if cfg.cache != nil {
						mu.Lock()
						cfg.cache.AddResult(site)
						mu.Unlock()
					}
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
				if cfg.cache != nil {
					cfg.cache.AddResult(site)
				}
				mu.Unlock()

				// Add to interrupt data for Ctrl+C handling
				if cfg.interruptData != nil {
					cfg.interruptData.AddWebsite(site)
				}

				if cfg.domainTitle != "" && len(site) > 2 && strings.EqualFold(site[2], cfg.domainTitle) && !cfg.con {
					mu.Lock()
					stopped = true
					mu.Unlock()
					return
				}
			}

			// Only update progress bar if not cancelled
			if cfg.interruptData == nil || !cfg.interruptData.IsCancelled() {
				mu.Lock()
				_ = bar.Add(1)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	_ = bar.Finish()

	// Batch DNS lookup for all found sites (much faster than individual lookups)
	if len(foundSites) > 0 && !stopped && (cfg.interruptData == nil || !cfg.interruptData.IsCancelled()) {
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
					// Append hostname to result
					if len(site) == 3 {
						foundSites[i] = append(site, hostname)
					}
				}
			}
		}

		// Also update interrupt data
		if cfg.interruptData != nil {
			cfg.interruptData.mu.Lock()
			for i, site := range cfg.interruptData.Websites {
				if len(site) >= 2 {
					lookupIP := strings.TrimPrefix(strings.TrimPrefix(site[1], "https://"), "http://")
					if hostname, ok := dnsResults[lookupIP]; ok && hostname != "" {
						if len(site) == 3 {
							cfg.interruptData.Websites[i] = append(site, hostname)
						}
					}
				}
			}
			cfg.interruptData.mu.Unlock()
		}

		fmt.Printf("[*] DNS lookup completed: %d/%d hostnames resolved\n", len(dnsResults), len(ipsToLookup))
	}

	// Print scan statistics
	fmt.Printf("\n[*] Scan Statistics: %d/%d IPs scanned, %d sites found\n", scannedCount, len(cfg.ipAddress), foundCount)

	// Process and print results (sync.Once guarantees single execution)
	printOnce.Do(func() {
		mu.Lock()
		wasEarlyStopped := stopped
		mu.Unlock()
		_ = bar.Finish()
		if cfg.cache != nil {
			cfg.cache.MarkCompleted()
			_ = cfg.cache.Save()
			config.InfoLog("Cache saved to: %s", cfg.cache.FilePath)
		}
		if wasEarlyStopped {
			PrintResult("Search Domain by ASN"+cfg.methodPrefix, cfg.domainTitle, cfg.timeout, cfg.ipBlocks, foundSites, cfg.export)
		} else {
			PrintResult("Search All ASN/IP"+cfg.methodPrefix, cfg.domainTitle, cfg.timeout, cfg.ipBlocks, foundSites, cfg.export)
		}
	})
}

// ResolveSite performs parallel scanning of IP addresses for websites
func ResolveSite(IPAddress []string, DomainTitle string, IPBlocks []string, domain string, con bool, export bool, timeout int, interruptData *InterruptData) {
	runScan(&scanConfig{
		ipAddress:     IPAddress,
		domainTitle:   DomainTitle,
		ipBlocks:      IPBlocks,
		domain:        domain,
		con:           con,
		export:        export,
		timeout:       timeout,
		interruptData: interruptData,
		cache:         nil,
		methodPrefix:  "",
	})
}

// ResolveSiteWithCache performs scanning with cache support for resuming
func ResolveSiteWithCache(IPAddress []string, DomainTitle string, IPBlocks []string, domain string, con bool, export bool, timeout int, interruptData *InterruptData, cache *Cache) {
	runScan(&scanConfig{
		ipAddress:     IPAddress,
		domainTitle:   DomainTitle,
		ipBlocks:      IPBlocks,
		domain:        domain,
		con:           con,
		export:        export,
		timeout:       timeout,
		interruptData: interruptData,
		cache:         cache,
		methodPrefix:  " (Resumed)",
	})
}
