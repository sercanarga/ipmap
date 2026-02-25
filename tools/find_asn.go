package tools

import (
	"fmt"
	"ipmap/config"
	"ipmap/modules"
	"regexp"
	"strconv"
	"time"
)

func FindASN(asn string, domain string, domainTitle string, con bool, export bool, timeout int, interruptData *modules.InterruptData) {
	// Use local variables instead of global to avoid race conditions
	var ipBlocks []string
	var ipAddress []string

	// Match both IPv4 (route:) and IPv6 (route6:) blocks
	re := regexp.MustCompile(`(?m)route6?:\s+([0-9a-fA-F:\./]+)$`)
	for _, match := range re.FindAllStringSubmatch(modules.FindIPBlocks(asn), -1) {
		ipBlocks = append(ipBlocks, match[1])
	}

	if len(ipBlocks) == 0 {
		config.ErrorLog("No IP blocks found for ASN: %s", asn)
		return
	}

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

	// Update interrupt data with IP blocks
	if interruptData != nil {
		interruptData.IPBlocks = ipBlocks
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

	fmt.Println("ASN:         " + asn +
		"\nIP Block:    " + strconv.Itoa(len(ipBlocks)) +
		"\nIP Address:  " + strconv.Itoa(len(ipAddress)) +
		"\nWorkers:     " + strconv.Itoa(workerCount) +
		"\nStart Time:  " + time.Now().Local().Format("2006-01-02 15:04:05") +
		"\nEst. End:    " + time.Now().Add(time.Duration(estimatedSeconds)*time.Second).Local().Format("2006-01-02 15:04:05"))

	modules.ResolveSite(ipAddress, domainTitle, ipBlocks, domain, con, export, timeout, interruptData)
}
