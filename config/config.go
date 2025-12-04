package config

import "fmt"

var (
	Verbose    bool
	Format     string
	MaxRetries int = 2   // Default retry count
	Workers    int = 100 // Default concurrent workers
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
