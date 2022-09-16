package modules

import "fmt"

func ResolveSite(IPAddress []string, Websites [][]string, DomainTitle string, IPBlocks []string, domain string, con bool, export bool, timeout int) {
	for _, ip := range IPAddress {
		site := GetSite(ip, domain, timeout)
		if len(site) > 0 {
			fmt.Print("+")
			Websites = append(Websites, site)

			if DomainTitle != "" && site[2] == DomainTitle && con == false {
				PrintResult("Search Domain by ASN", DomainTitle, timeout, IPBlocks, Websites, export)
				return
			}
		} else {
			fmt.Print("-")
		}
	}

	PrintResult("Search All ASN/IP", DomainTitle, timeout, IPBlocks, Websites, export)
}
