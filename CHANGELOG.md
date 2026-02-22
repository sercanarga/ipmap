# Changelog

All notable changes to this project will be documented in this file.

## [2.4.0] - 2026-02-22

### Added
- **Go Embed Version**: VERSION file embedded at compile time via `go:embed`
- **Domain Validation**: `-d` flag now validates domain format before processing
- **Backward-Compatible JSON**: `found_websites` + `founded_websites` dual output
- **Config File Support**: `-config config.yaml` flag for YAML configuration
- **Resume Support**: `-resume cache.json` flag to resume interrupted scans
- **Output Directory**: `-output-dir ./exports` flag for export file location

### Fixed
- **Race Condition**: `PrintResult` protected with `sync.Once` (prevents double-call)
- **DNS Key Mismatch**: Protocol prefix stripped before DNS result lookup
- **CLI Flag Override**: `flag.Visit` ensures only explicitly set flags override config
- **Windows Path Separator**: `filepath.Join` replaces hardcoded `/`
- **Flaky DNS Test**: `minExpected` set to 0 for network-dependent test
- **Chrome 131 Comment**: Updated stale comment to Chrome 135

### Improved
- **Cache Performance**: `IsScanned()` O(n) → O(1) with persistent `map[string]struct{}`
- **JSON Field Name**: `founded_websites` → `found_websites` (grammar fix)

### Removed
- Unused `GetTransport()` method (conflicting transport settings)
- Unused `AddSmartJitter()`, `SaveConfigFile()`, `ParseDNSServers()` functions
- Unused `BatchReverseDNSWithResults()` function and `DNSResult` struct
- Unused `PoolMetrics` struct and related atomic operations

---

## [2.3.0] - 2025-12-23

### Added
- **Batch DNS Lookup**: Parallel DNS resolution at end of scan (20 concurrent queries)
- **Connection Pool Optimization**: Pre-warming, larger buffers, TLS session cache

### Changed
- **Chrome 135 User-Agent**: Updated from Chrome 131 to Chrome 135 for January 2026
- **macOS 15 Sequoia**: Added new macOS version strings
- **Windows 11 24H2**: Updated Windows version strings
- Increased connection pool sizes: MaxIdle 200-1000, MaxPerHost 50-200
- Extended idle timeout: 60s → 90s
- Larger I/O buffers: 64KB read/write

### Fixed
- Batch DNS now correctly extracts IP from URL (removes https:// prefix)

---

## [2.2.2] - 2025-12-12

### Fixed
- **uTLS Transport Activated**: Chrome 131 TLS fingerprint now fully integrated into HTTP client
- Fixed HTTP/2 compatibility issue by setting `ForceAttemptHTTP2: false`
- uTLS transport now properly used for all HTTPS connections

---

## [2.2.1] - 2025-12-11

### Added
- **Referer Header Rotation**: Random referer from Google, Bing, DuckDuckGo for more realistic requests
- **Smart Jitter Function**: `config.AddSmartJitter()` *(removed in v2.4.0)*
- **uTLS Transport**: Chrome 131 TLS fingerprint support
- New config constants: `DialTimeout`, `MaxJitterMs`
- Unified RNG functions in config: `GetRandomInt`, `GetRandomString`, `ShuffleStrings`

### Changed
- Reduced jitter from 100-800ms to 0-200ms for faster scanning
- Reduced timeouts for better performance:
  - TLS handshake: 10s → 5s
  - Response header: 10s → 5s
  - Dial timeout: 10s → 5s
  - Retry backoff: 500ms → 200ms
- `ValidateTimeout` now uses `config.MinTimeout` constant
- Updated GitHub URL to `sercanarga/ipmap`

### Fixed
- Domain resolution failure with uTLS HTTP/2 compatibility
- Hardcoded timeout values now use config constants

### Removed
- ~500 lines of dead code from scanner.go
- Unused uTLS dependencies (temporarily, readded for future use)

## [2.0.0] - 2025-12-10

### Added
- Chrome 131 TLS fingerprint (JA3/JA4 spoofing)
- Real Chrome header order for WAF bypass
- Proxy support (HTTP/HTTPS/SOCKS5)
- Custom DNS servers
- Rate limiting (token bucket algorithm)
- IP shuffling for firewall bypass
- Graceful Ctrl+C handling with export option
- Input validation for ASN, IP/CIDR formats
- Verbose logging mode
- JSON output format

### Changed
- Complete rewrite of HTTP client
- Improved concurrent worker management
- Better error handling and recovery

---

## Version History
- **2.1.0**: Anti-detection improvements, performance optimization
- **2.0.0**: Major rewrite with anti-detection features
- **1.0.0**: Initial release
