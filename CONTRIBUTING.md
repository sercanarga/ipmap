# Contributing to ipmap

Thank you for your interest in contributing to ipmap! We welcome contributions from the community.

## How to Contribute

### Reporting Bugs
- Check if the bug has already been reported in [Issues](https://github.com/sercanarga/ipmap/issues)
- Provide a clear description of the bug
- Include steps to reproduce the issue
- Attach relevant logs or error messages

### Suggesting Enhancements
- Check if the enhancement has already been suggested
- Provide a clear description of the proposed feature
- Explain the use case and benefits

### Pull Requests
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add or update tests as needed
5. Ensure all tests pass (`go test ./... -v`)
6. Run `go vet ./...` to check for issues
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## Development Setup

```bash
# Clone the repository
git clone https://github.com/sercanarga/ipmap.git
cd ipmap

# Install dependencies
go mod tidy

# Build the project
go build -o ipmap .

# Run tests
go test ./... -v

# Run tests with coverage
go test ./... -v -cover

# Run static analysis
go vet ./...

# Run with verbose output
./ipmap -asn AS13335 -v
```

## Project Structure

```
ipmap/
├── main.go              # Entry point, CLI flags, go:embed VERSION
├── VERSION              # Embedded version string
├── config/
│   ├── config.go        # Global config, RNG, jitter, logging
│   └── loader.go        # YAML config file loading
├── modules/
│   ├── scanner.go       # Chrome 135 headers, uTLS transport
│   ├── request.go       # HTTP client with retry, connection pool
│   ├── resolve_site.go  # Worker pool, IP scanning, batch DNS
│   ├── get_site.go      # Site discovery per IP
│   ├── get_domain_title.go # Domain title fetching
│   ├── helpers.go       # Shared utilities (ExtractTitle)
│   ├── cache.go         # Scan state persistence for resume
│   ├── validators.go    # Input validation
│   ├── rate_limiter.go  # Token bucket rate limiter
│   ├── dns_resolver.go  # Batch reverse DNS lookups
│   ├── result_print.go  # Result formatting and export
│   ├── interrupt_handler.go # Ctrl+C handling
│   └── calc_ip_address.go   # CIDR to IP calculation
├── tools/
│   ├── find_asn.go      # ASN scanning
│   └── find_ip.go       # IP block scanning
└── README.md
```

## Code Style

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and concise
- Run `go fmt` before committing
- Run `go vet` to catch common issues

## Testing

- Write tests for new features
- Ensure all existing tests pass
- Aim for good test coverage (current: ~30%)
- Test edge cases and error conditions
- Place tests in `*_test.go` files

### Running Tests

```bash
# All tests
go test ./... -v

# Specific package
go test ./modules -v

# With coverage report
go test ./... -v -cover

# Run benchmarks
go test ./... -bench=.
```

## Anti-Detection Guidelines

When modifying the scanner module:

1. **TLS Fingerprint**: Use `utls.HelloChrome_Auto` for latest Chrome fingerprint
2. **Header Order**: Maintain exact Chrome header order (not alphabetical)
3. **Accept-Encoding**: Include `zstd` for Chrome 135+
4. **Jitter**: Use `config.AddJitter()` (0-200ms random delay)
5. **User-Agent**: Use Chrome 133+ versions only
6. **Referer**: Rotate between Google, Bing, DuckDuckGo, Yahoo URLs

## License

By contributing to ipmap, you agree that your contributions will be licensed under the MIT License.

