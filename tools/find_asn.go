package tools

import (
	"fmt"
	"ipmap/modules"
	"regexp"
	"strconv"
	"time"
)

func FindASN(asn string, domain string, domainTitle string, con bool, export bool, timeout int, interruptData *modules.InterruptData) {
	// Use local variables instead of global to avoid race conditions
	var ipBlocks []string
	var ipAddress []string
	var websites [][]string

	re := regexp.MustCompile(`(?m)route:\s+([0-9\.\/]+)$`)
	for _, match := range re.FindAllStringSubmatch(modules.FindIPBlocks(asn), -1) {
		ipBlocks = append(ipBlocks, match[1])
	}

	for _, block := range ipBlocks {
		ips, err := modules.CalcIPAddress(block)
		if err != nil {
			return
		}

		ipAddress = append(ipAddress, ips...)
	}

	// Update interrupt data with IP blocks
	if interruptData != nil {
		interruptData.IPBlocks = ipBlocks
	}

	fmt.Println("ASN:         " + asn +
		"\nIP Block:    " + strconv.Itoa(len(ipBlocks)) +
		"\nIP Address:  " + strconv.Itoa(len(ipAddress)) +
		"\nStart Time:  " + time.Now().Local().String() +
		"\nEnd Time:    " + time.Now().Add((time.Millisecond*time.Duration(timeout))*time.Duration(len(ipAddress))).Local().String())

	modules.ResolveSite(ipAddress, websites, domainTitle, ipBlocks, domain, con, export, timeout, interruptData)
}
