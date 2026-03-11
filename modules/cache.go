// cache.go provides result caching for interrupted scans
// Allows resuming scans from where they left off
package modules

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"
)

// CacheData represents the cached scan state
type CacheData struct {
	ASN         string     `json:"asn,omitempty"`
	Domain      string     `json:"domain,omitempty"`
	DomainTitle string     `json:"domain_title,omitempty"`
	Timeout     int        `json:"timeout"`
	IPBlocks    []string   `json:"ip_blocks"`
	ScannedIPs  []string   `json:"scanned_ips"`
	Results     [][]string `json:"results"`
	LastUpdate  string     `json:"last_update"`
	Completed   bool       `json:"completed"`
}

// Cache manages scan state persistence
type Cache struct {
	Data       CacheData
	FilePath   string
	scannedSet map[string]struct{} // O(1) lookup for scanned IPs
	mu         sync.RWMutex
}

// NewCache creates a new cache instance
func NewCache(filePath string) *Cache {
	return &Cache{
		FilePath:   filePath,
		scannedSet: make(map[string]struct{}),
		Data: CacheData{
			ScannedIPs: make([]string, 0),
			Results:    make([][]string, 0),
		},
	}
}

// LoadCache loads cache from file
func LoadCache(filePath string) (*Cache, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, err
	}

	// Build scanned set from loaded data for O(1) lookups
	scannedSet := make(map[string]struct{}, len(cacheData.ScannedIPs))
	for _, ip := range cacheData.ScannedIPs {
		scannedSet[ip] = struct{}{}
	}

	return &Cache{
		Data:       cacheData,
		FilePath:   filePath,
		scannedSet: scannedSet,
	}, nil
}

// Save persists cache to file
func (c *Cache) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Data.LastUpdate = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(c.Data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.FilePath, data, 0644)
}

// AddScannedIP marks an IP as scanned
func (c *Cache) AddScannedIP(ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.ScannedIPs = append(c.Data.ScannedIPs, ip)
	c.scannedSet[ip] = struct{}{}
}

// AddResult adds a scan result
func (c *Cache) AddResult(result []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Results = append(c.Data.Results, result)
}

// IsScanned checks if an IP was already scanned (O(1) via map lookup)
func (c *Cache) IsScanned(ip string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.scannedSet[ip]
	return exists
}

// GetUnscannedIPs returns IPs that haven't been scanned yet (O(1) per IP via map)
func (c *Cache) GetUnscannedIPs(allIPs []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Use persistent scannedSet for O(1) lookup
	unscanned := make([]string, 0, len(allIPs)-len(c.scannedSet))
	for _, ip := range allIPs {
		if _, exists := c.scannedSet[ip]; !exists {
			unscanned = append(unscanned, ip)
		}
	}

	return unscanned
}

// SetMetadata sets scan metadata
func (c *Cache) SetMetadata(asn, domain, domainTitle string, timeout int, ipBlocks []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Data.ASN = asn
	c.Data.Domain = domain
	c.Data.DomainTitle = domainTitle
	c.Data.Timeout = timeout
	c.Data.IPBlocks = ipBlocks
}

// MarkCompleted marks the scan as completed
func (c *Cache) MarkCompleted() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Completed = true
}

// GetResults returns all cached results
func (c *Cache) GetResults() [][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Data.Results
}

// GetProgress returns scan progress information
func (c *Cache) GetProgress() (scanned int, total int, results int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	totalIPs := len(c.Data.ScannedIPs)
	// Calculate actual total from IP blocks using CIDR parsing
	if len(c.Data.IPBlocks) > 0 {
		totalIPs = 0
		for _, block := range c.Data.IPBlocks {
			ips, err := CalcIPAddress(block)
			if err != nil {
				// Fallback to rough estimate if CIDR parsing fails
				totalIPs += 256
				continue
			}
			totalIPs += len(ips)
		}
	}
	return len(c.Data.ScannedIPs), totalIPs, len(c.Data.Results)
}

// GenerateCacheFileName generates a cache file name based on scan parameters
func GenerateCacheFileName(asn, domain string) string {
	if asn != "" {
		return "ipmap_" + asn + "_cache.json"
	}
	if domain != "" {
		return "ipmap_" + SanitizeFilename(domain) + "_cache.json"
	}
	return "ipmap_" + strconv.FormatInt(time.Now().Unix(), 10) + "_cache.json"
}
