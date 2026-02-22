package tools

import (
	"fmt"
	"ipmap/config"
	"ipmap/modules"
	"strconv"
	"time"
)

func FindIP(ipBlocks []string, domain string, domainTitle string, con bool, export bool, timeout int, interruptData *modules.InterruptData) {
	// Use local variables instead of global to avoid race conditions
	var ipAddress []string
	var websites [][]string

	for _, block := range ipBlocks {
		ips, err := modules.CalcIPAddress(block)
		if err != nil {
			config.ErrorLog("Failed to parse CIDR block '%s': %v", block, err)
			continue // Continue with other blocks instead of returning
		}

		ipAddress = append(ipAddress, ips...)
	}

	if len(ipAddress) == 0 {
		config.ErrorLog("No valid IP addresses to scan")
		return
	}

	// Calculate estimated end time based on workers (parallel processing)
	workerCount := config.Workers
	if workerCount <= 0 {
		workerCount = 100
	}
	estimatedSeconds := (len(ipAddress) / workerCount) * timeout / 1000
	if estimatedSeconds < 1 {
		estimatedSeconds = 1
	}

	fmt.Println("IP Block:    " + strconv.Itoa(len(ipBlocks)) +
		"\nIP Address:  " + strconv.Itoa(len(ipAddress)) +
		"\nWorkers:     " + strconv.Itoa(workerCount) +
		"\nStart Time:  " + time.Now().Local().Format("2006-01-02 15:04:05") +
		"\nEst. End:    " + time.Now().Add(time.Duration(estimatedSeconds)*time.Second).Local().Format("2006-01-02 15:04:05"))

	modules.ResolveSite(ipAddress, websites, domainTitle, ipBlocks, domain, con, export, timeout, interruptData)
}

// FindIPWithCache resumes a scan from cache, filtering out already-scanned IPs
func FindIPWithCache(ipBlocks []string, domain string, domainTitle string, con bool, export bool, timeout int, interruptData *modules.InterruptData, cache *modules.Cache) {
	var ipAddress []string
	var websites [][]string

	for _, block := range ipBlocks {
		ips, err := modules.CalcIPAddress(block)
		if err != nil {
			config.ErrorLog("Failed to parse CIDR block '%s': %v", block, err)
			continue
		}

		ipAddress = append(ipAddress, ips...)
	}

	if len(ipAddress) == 0 {
		config.ErrorLog("No valid IP addresses to scan")
		return
	}

	// Filter out already-scanned IPs from cache
	unscannedIPs := cache.GetUnscannedIPs(ipAddress)

	config.InfoLog("Total IPs: %d, Already scanned: %d, Remaining: %d",
		len(ipAddress), len(ipAddress)-len(unscannedIPs), len(unscannedIPs))

	if len(unscannedIPs) == 0 {
		config.InfoLog("All IPs already scanned, loading results from cache")
		modules.PrintResult("Search All ASN/IP (Cached)", domainTitle, timeout, ipBlocks, cache.GetResults(), export)
		return
	}

	// Calculate estimated end time
	workerCount := config.Workers
	if workerCount <= 0 {
		workerCount = 100
	}
	estimatedSeconds := (len(unscannedIPs) / workerCount) * timeout / 1000
	if estimatedSeconds < 1 {
		estimatedSeconds = 1
	}

	fmt.Println("Resuming Scan:" +
		"\nIP Block:    " + strconv.Itoa(len(ipBlocks)) +
		"\nTotal IPs:   " + strconv.Itoa(len(ipAddress)) +
		"\nRemaining:   " + strconv.Itoa(len(unscannedIPs)) +
		"\nWorkers:     " + strconv.Itoa(workerCount) +
		"\nEst. End:    " + time.Now().Add(time.Duration(estimatedSeconds)*time.Second).Local().Format("2006-01-02 15:04:05"))

	modules.ResolveSiteWithCache(unscannedIPs, websites, domainTitle, ipBlocks, domain, con, export, timeout, interruptData, cache)
}
