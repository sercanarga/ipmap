    package modules

    import (
        "fmt"
        "sync"
    )

    func ResolveSite(IPAddress []string, Websites [][]string, DomainTitle string, IPBlocks []string, domain string, con bool, export bool, timeout int) {
        var wg sync.WaitGroup
        var mu sync.Mutex
        sem := make(chan struct{}, 100)

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

                    if DomainTitle != "" && site[2] == DomainTitle && con == false {
                        PrintResult("Search Domain by ASN", DomainTitle, timeout, IPBlocks, Websites, export)
                        return
                    }
                } else {
                    fmt.Print("-")
                }
            }(ip)
        }

        wg.Wait()

        // Sonuçları toplu olarak işlemek için PrintResult çağrısı
        PrintResult("Search All ASN/IP", DomainTitle, timeout, IPBlocks, Websites, export)
    }
