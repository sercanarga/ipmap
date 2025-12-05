package config

import "fmt"

var (
	Verbose    bool
	Format     string
	MaxRetries int = 2   // Default retry count
	Workers    int = 100 // Default concurrent workers

	// New features
	ProxyURL   string       // HTTP/HTTPS/SOCKS5 proxy URL
	RateLimit  int      = 0 // Requests per second (0 = unlimited)
	DNSServers []string     // Custom DNS servers
)

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
