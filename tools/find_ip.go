package tools

import (
	"fmt"
	"ipmap/modules"
	"strconv"
	"time"
)

func FindIP(IPBlocks []string, domain string, domainTitle string, con bool, export bool, timeout int) {
	for _, block := range IPBlocks {
		ips, err := modules.CalcIPAddress(block)
		if err != nil {
			return
		}

		IPAddress = append(IPAddress, ips...)
	}

	fmt.Println("IP Block:    " + strconv.Itoa(len(IPBlocks)) +
		"\nIP Address:  " + strconv.Itoa(len(IPAddress)) +
		"\nStart Time:  " + time.Now().Local().String() +
		"\nEnd Time:    " + time.Now().Add((time.Millisecond*time.Duration(timeout))*time.Duration(len(IPAddress))).Local().String())

	modules.ResolveSite(IPAddress, Websites, domainTitle, IPBlocks, domain, con, export, timeout)
}
