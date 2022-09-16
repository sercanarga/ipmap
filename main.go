package main

import (
	"flag"
	"fmt"
	"ipmap/modules"
	"ipmap/tools"
	"strconv"
	"strings"
)

var (
	domain      = flag.String("d", "", "domain parameter")
	asn         = flag.String("asn", "", "asn parameter")
	ip          = flag.String("ip", "", "ip parameter")
	timeout     = flag.Int("t", 0, "timeout parameter")
	con         = flag.Bool("c", false, "continue parameter")
	export      = flag.Bool("export", false, "export parameter")
	DomainTitle string
)

func main() {
	flag.Parse()
	if (*asn != "" && *ip != "") || (*asn == "" && *ip == "") {
		fmt.Println("======================================================\n" +
			"      ipmap v1.0 (github.com/sercanarga/ipmap)\n" +
			"======================================================\n" +
			"PARAMETERS:\n" +
			"-asn AS13335\n" +
			"-ip 103.21.244.0/22,103.22.200.0/22\n" +
			"-d example.com\n" +
			"-t 200 (timout default:auto)\n" +
			"--c (work until finish scanning)\n" +
			"--export (auto export results)\n\n" +
			"USAGES:\n" +
			"Finding sites by scanning all the IP blocks\nipmap -ip 103.21.244.0/22,103.22.200.0/22\n\n" +
			"Finding real IP address of site by scanning given IP addresses\nipmap -ip 103.21.244.0/22,103.22.200.0/22 -d example.com\n\n" +
			"Finding sites by scanning all the IP blocks in the ASN\nipmap -asn AS13335\n\n" +
			"Finding real IP address of site by scanning all IP blocks in ASN\nipmap -asn AS13335 -d example.com")
		return
	}

	if *timeout == 0 && *domain == "" {
		fmt.Println("Timeout parameter( -t ) is not set. By entering the domain, you can have it calculated automatically.")
		return
	}

	if *domain != "" {
		getDomain := modules.GetDomainTitle(*domain)
		if len(getDomain) == 0 {
			fmt.Println("Domain not resolved.")
			return
		}
		DomainTitle = getDomain[0]

		if *timeout == 0 {
			resolveTime, _ := strconv.Atoi(getDomain[1])
			*timeout = ((resolveTime * 15) / 100) + resolveTime
		}
	}

	if *ip != "" {
		splitIP := strings.Split(*ip, ",")
		tools.FindIP(splitIP, *domain, DomainTitle, *con, *export, *timeout)
		return
	}

	if *asn != "" {
		tools.FindASN(*asn, *domain, DomainTitle, *con, *export, *timeout)
		return
	}

}
