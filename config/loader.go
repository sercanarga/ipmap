// Package config provides global configuration for the ipmap scanner.
// loader.go handles configuration file loading and parsing.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the structure of the config.yaml file
type FileConfig struct {
	Workers    int      `yaml:"workers"`
	Timeout    int      `yaml:"timeout"`
	RateLimit  int      `yaml:"rate_limit"`
	Proxy      string   `yaml:"proxy"`
	DNSServers []string `yaml:"dns_servers"`
	IPv6       bool     `yaml:"ipv6"`
	Verbose    bool     `yaml:"verbose"`
	Format     string   `yaml:"format"`
}

// LoadConfigFile loads configuration from a YAML file
// Returns nil if file doesn't exist or is invalid
func LoadConfigFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ApplyFileConfig applies file configuration to global config
// Only applies non-zero/non-empty values (allows CLI to override)
func ApplyFileConfig(cfg *FileConfig) {
	if cfg == nil {
		return
	}

	if cfg.Workers > 0 {
		Workers = cfg.Workers
	}
	if cfg.RateLimit > 0 {
		RateLimit = cfg.RateLimit
	}
	if cfg.Proxy != "" {
		ProxyURL = cfg.Proxy
	}
	if len(cfg.DNSServers) > 0 {
		DNSServers = cfg.DNSServers
	}
	if cfg.IPv6 {
		EnableIPv6 = cfg.IPv6
	}
	if cfg.Verbose {
		Verbose = cfg.Verbose
	}
	if cfg.Format != "" {
		Format = cfg.Format
	}
}

// FindConfigFile looks for config file in common locations
func FindConfigFile() string {
	// Check common locations in order
	locations := []string{
		"config.yaml",
		"config.yml",
		".ipmap.yaml",
		".ipmap.yml",
	}

	// Also check in user home directory
	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations, home+"/.ipmap.yaml", home+"/.ipmap.yml")
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}
