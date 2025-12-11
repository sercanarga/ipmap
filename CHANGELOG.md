# Changelog

All notable changes to this project will be documented in this file.

## [2.2.1] - 2025-12-11

### Added
- **Referer Header Rotation**: Random referer from Google, Bing, DuckDuckGo for more realistic requests
- **Smart Jitter Function**: `config.AddSmartJitter()` with occasional long pauses (1-3s) for natural patterns
- **uTLS Transport**: Chrome 131 TLS fingerprint support (ready for integration)
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
- Updated GitHub URL to `lordixir/ipmap`

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
