package main

import (
	"flag"
	"fmt"
	"ipmap/config"
	"ipmap/modules"
	"ipmap/tools"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	domain      = flag.String("d", "", "domain parameter")
	asn         = flag.String("asn", "", "asn parameter")
	ip          = flag.String("ip", "", "ip parameter")
	timeout     = flag.Int("t", 0, "timeout parameter")
	con         = flag.Bool("c", false, "continue parameter")
	export      = flag.Bool("export", false, "export parameter")
	verbose     = flag.Bool("v", false, "verbose mode")
	format      = flag.String("format", "text", "output format (text/json)")
	workers     = flag.Int("workers", 100, "number of concurrent workers")
	proxy       = flag.String("proxy", "", "proxy URL (http/https/socks5)")
	rate        = flag.Int("rate", 0, "requests per second (0 = unlimited)")
	dns         = flag.String("dns", "", "custom DNS servers (comma-separated)")
	DomainTitle string

	// Global state for interrupt handling
	interruptData *modules.InterruptData
)

func main() {
	flag.Parse()

	// Set global config
	config.Verbose = *verbose
	config.Format = *format
	config.Workers = modules.ValidateWorkerCount(*workers)
	config.ProxyURL = *proxy
	config.RateLimit = *rate
	if *dns != "" {
		config.DNSServers = strings.Split(*dns, ",")
	}

	// Setup interrupt handler
	interruptData = &modules.InterruptData{}
	setupInterruptHandler()

	// Log configuration if verbose
	if config.Verbose {
		config.VerboseLog("Configuration - Workers: %d, Rate Limit: %d/s, Proxy: %s",
			config.Workers, config.RateLimit, config.ProxyURL)
		if len(config.DNSServers) > 0 {
			config.VerboseLog("Custom DNS Servers: %v", config.DNSServers)
		}
	}

	if (*asn != "" && *ip != "") || (*asn == "" && *ip == "") {
		fmt.Println("======================================================\n" +
			"      ipmap v2.0 (github.com/sercanarga/ipmap)\n" +
			"======================================================\n" +
			"PARAMETERS:\n" +
			"-asn AS13335\n" +
			"-ip 103.21.244.0/22,103.22.200.0/22\n" +
			"-d example.com\n" +
			"-t 200 (timeout default:auto)\n" +
			"-c (work until finish scanning)\n" +
			"--export (auto export results)\n" +
			"-v (verbose mode)\n" +
			"-format json (output format: text/json)\n" +
			"-workers 100 (concurrent workers, default: 100)\n" +
			"-proxy http://127.0.0.1:8080 (proxy URL)\n" +
			"-rate 50 (requests per second, 0 = unlimited)\n" +
			"-dns 8.8.8.8,1.1.1.1 (custom DNS servers)\n\n" +
			"USAGES:\n" +
			"Finding sites by scanning all the IP blocks\nipmap -ip 103.21.244.0/22,103.22.200.0/22\n\n" +
			"Finding real IP address of site by scanning given IP addresses\nipmap -ip 103.21.244.0/22,103.22.200.0/22 -d example.com\n\n" +
			"Finding sites by scanning all the IP blocks in the ASN\nipmap -asn AS13335\n\n" +
			"Finding real IP address of site by scanning all IP blocks in ASN\nipmap -asn AS13335 -d example.com\n\n" +
			"Using proxy and rate limiting\nipmap -asn AS13335 -proxy http://127.0.0.1:8080 -rate 50")
		return
	}

	if *timeout == 0 && *domain == "" {
		fmt.Println("Timeout parameter( -t ) is not set. By entering the domain, you can have it calculated automatically.")
		return
	}

	if *domain != "" {
		getDomain := modules.GetDomainTitle(*domain)
		if len(getDomain) == 0 {
			fmt.Println("Domain not resolved. Please check:")
			fmt.Println("  - Domain is accessible via HTTP/HTTPS")
			fmt.Println("  - No network/firewall issues")
			fmt.Println("  - Domain name is correct")
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
		interruptData.IPBlocks = splitIP
		interruptData.Domain = DomainTitle
		interruptData.Timeout = *timeout
		tools.FindIP(splitIP, *domain, DomainTitle, *con, *export, *timeout, interruptData)
		return
	}

	if *asn != "" {
		interruptData.Domain = DomainTitle
		interruptData.Timeout = *timeout
		tools.FindASN(*asn, *domain, DomainTitle, *con, *export, *timeout, interruptData)
		return
	}

}

func setupInterruptHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n[!] Scan interrupted by user")

		if interruptData != nil && len(interruptData.Websites) > 0 {
			fmt.Printf("\n[*] Found %d websites before interruption\n", len(interruptData.Websites))
			fmt.Print("\nDo you want to export the results? (Y/n): ")

			var response string
			_, _ = fmt.Scanln(&response)

			if response == "y" || response == "Y" || response == "" {
				modules.PrintResult("Search Interrupted", interruptData.Domain, interruptData.Timeout,
					interruptData.IPBlocks, interruptData.Websites, true)
				fmt.Println("\n[✓] Results exported successfully")
			} else {
				fmt.Println("\n[✗] Export canceled")
			}
		} else {
			fmt.Println("\n[!] No results to export")
		}

		os.Exit(0)
	}()
}
