// Package config provides global configuration for the ipmap scanner.
// It includes timeout settings, worker counts, and network options.
package config

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Timeout constants (in milliseconds)
const (
	// DefaultDomainTimeout is the timeout for domain resolution requests
	DefaultDomainTimeout = 15000

	// DefaultAPITimeout is the timeout for external API calls (RADB, etc.)
	DefaultAPITimeout = 5000

	// DefaultBaseTimeout is the base timeout for IP scanning
	DefaultBaseTimeout = 2000

	// MaxTimeout is the maximum allowed timeout
	MaxTimeout = 10000

	// MinTimeout is the minimum allowed timeout
	MinTimeout = 100

	// DialTimeout is the TCP connection timeout
	DialTimeout = 5

	// MaxJitterMs is the maximum jitter delay in milliseconds
	MaxJitterMs = 200
)

// Global configuration variables
var (
	// Verbose enables detailed logging output
	Verbose bool

	// Format specifies output format ("text" or "json")
	Format string

	// MaxRetries is the number of retry attempts for failed requests
	MaxRetries int = 0

	// Workers is the number of concurrent scanning goroutines
	Workers int = 100

	// ProxyURL is the HTTP/HTTPS/SOCKS5 proxy URL
	ProxyURL string

	// RateLimit is requests per second (0 = unlimited)
	RateLimit int = 0

	// DNSServers is the list of custom DNS servers
	DNSServers []string

	// EnableIPv6 enables IPv6 address scanning (default: false)
	EnableIPv6 bool = false

	// OutputDir is the directory for export files (default: current directory)
	OutputDir string = ""

	// InsecureSkipVerify skips TLS certificate verification (default: true for backward compatibility)
	InsecureSkipVerify bool = true
)

// ====================================================================
// UNIFIED RANDOM GENERATOR (thread-safe)
// ====================================================================

var (
	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
	rngMu sync.Mutex
)

// GetRandomInt returns a random int in range [0, max)
func GetRandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	rngMu.Lock()
	defer rngMu.Unlock()
	return rng.Intn(max)
}

// GetRandomString returns a random string from the given slice
func GetRandomString(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[GetRandomInt(len(slice))]
}

// ShuffleStrings randomizes the order of strings in a slice
func ShuffleStrings(items []string) []string {
	shuffled := make([]string, len(items))
	copy(shuffled, items)
	rngMu.Lock()
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	rngMu.Unlock()
	return shuffled
}

// AddJitter adds random delay between 0 and MaxJitterMs
func AddJitter() {
	jitterMs := GetRandomInt(MaxJitterMs)
	if jitterMs > 0 {
		time.Sleep(time.Duration(jitterMs) * time.Millisecond)
	}
}

// ====================================================================
// LOGGING FUNCTIONS
// ====================================================================

// VerboseLog prints message only if verbose mode is enabled
func VerboseLog(format string, args ...interface{}) {
	if Verbose {
		fmt.Printf("[VERBOSE] "+format+"\n", args...)
	}
}

// ErrorLog prints error messages
func ErrorLog(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

// InfoLog prints info messages
func InfoLog(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

// WarnLog prints warning messages
func WarnLog(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}
