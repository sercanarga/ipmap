package tools

import (
	"fmt"
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
			return
		}

		ipAddress = append(ipAddress, ips...)
	}

	fmt.Println("IP Block:    " + strconv.Itoa(len(ipBlocks)) +
		"\nIP Address:  " + strconv.Itoa(len(ipAddress)) +
		"\nStart Time:  " + time.Now().Local().String() +
		"\nEnd Time:    " + time.Now().Add((time.Millisecond*time.Duration(timeout))*time.Duration(len(ipAddress))).Local().String())

	modules.ResolveSite(ipAddress, websites, domainTitle, ipBlocks, domain, con, export, timeout, interruptData)
}
