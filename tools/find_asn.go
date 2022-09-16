package tools

import (
	"fmt"
	"ipmap/modules"
	"regexp"
	"strconv"
	"time"
)

var (
	IPBlocks  []string
	IPAddress []string
	Websites  [][]string
)

func FindASN(asn string, domain string, domainTitle string, con bool, export bool, timeout int) {
	re := regexp.MustCompile(`(?m)route:\s+([0-9\.\/]+)$`)
	for _, match := range re.FindAllStringSubmatch(modules.FindIPBlocks(asn), -1) {
		IPBlocks = append(IPBlocks, match[1])
	}

	for _, block := range IPBlocks {
		ips, err := modules.CalcIPAddress(block)
		if err != nil {
			return
		}

		IPAddress = append(IPAddress, ips...)
	}

	fmt.Println("ASN:         " + asn +
		"\nIP Block:    " + strconv.Itoa(len(IPBlocks)) +
		"\nIP Address:  " + strconv.Itoa(len(IPAddress)) +
		"\nStart Time:  " + time.Now().Local().String() +
		"\nEnd Time:    " + time.Now().Add((time.Millisecond*time.Duration(timeout))*time.Duration(len(IPAddress))).Local().String())

	modules.ResolveSite(IPAddress, Websites, domainTitle, IPBlocks, domain, con, export, timeout)
}
