package modules

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"ipmap/config"
	"sync"
)

func ResolveSite(IPAddress []string, Websites [][]string, DomainTitle string, IPBlocks []string, domain string, con bool, export bool, timeout int, interruptData *InterruptData) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Use configurable worker pool size
	workerCount := config.Workers
	config.VerboseLog("Starting scan with %d concurrent workers", workerCount)
	sem := make(chan struct{}, workerCount)

	// Create progress bar
	bar := progressbar.NewOptions(len(IPAddress),
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

	for _, ip := range IPAddress {
		wg.Add(1)
		sem <- struct{}{}

		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			site := GetSite(ip, domain, timeout)
			if len(site) > 0 {

				fmt.Println("\n", site)
				mu.Lock()
				Websites = append(Websites, site)
				mu.Unlock()

				// Add to interrupt data for Ctrl+C handling
				if interruptData != nil {
					interruptData.AddWebsite(site)
				}

				if DomainTitle != "" && site[2] == DomainTitle && con == false {
					bar.Finish()
					PrintResult("Search Domain by ASN", DomainTitle, timeout, IPBlocks, Websites, export)
					return
				}
			}

			mu.Lock()
			bar.Add(1)
			mu.Unlock()
		}(ip)
	}

	wg.Wait()
	bar.Finish()

	// Process and print results
	PrintResult("Search All ASN/IP", DomainTitle, timeout, IPBlocks, Websites, export)
}
