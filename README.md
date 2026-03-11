# ipmap

An open-source, cross-platform powerful network analysis tool for discovering websites hosted on specific IP addresses and ASN ranges.

## Features

### Scanning
- **ASN scanning** — Queries RIPE Stat API for all announced prefixes of an ASN, with automatic fallback to RADB if RIPE is unavailable
- **IP block scanning** — Direct CIDR notation input with support for multiple comma-separated blocks
- **HTTPS/HTTP automatic fallback** — Tries HTTPS first, seamlessly falls back to HTTP if connection fails
- **Batch reverse DNS** — Automatically resolves hostnames for all discovered IPs after scan completion
- **IPv4/IPv6 support** — IPv6 scanning can be enabled with the `-ipv6` flag

### Anti-Detection & Firewall Bypass
- **Chrome 135 TLS Fingerprint** — JA3/JA4 fingerprint spoofing via [uTLS](https://github.com/refraction-networking/utls), making requests indistinguishable from real Chrome 135 browser traffic
- **Real Chrome Header Order** — Headers are sent in the exact order Chrome 135 uses (captured from DevTools), which is checked by Cloudflare and other WAFs
- **User-Agent Rotation** — Pool of 17 Chrome 133-135 User-Agents across Windows, macOS, and Linux
- **Referer Header Rotation** — Randomly cycles through Google, Bing, DuckDuckGo, and Yahoo referer URLs
- **IP Shuffling** — Randomizes scan order to avoid sequential scanning patterns that trigger firewalls
- **Smart Jitter** — Adds random delay (0-200ms) between requests for natural traffic patterns
- **Rate Limiting** — Token bucket algorithm to control requests per second

### Performance
- **Concurrent workers** — Configurable from 1 to 1000 parallel goroutines (default: 100)
- **Connection pooling** — Optimized HTTP connection pool with keep-alive and buffer tuning
- **Dynamic timeout** — Auto-calculated based on domain response time or worker count
- **Proxy connection pre-warming** — Pre-establishes connections for lower initial latency

### Resilience
- **Scan resume** — Interrupted scans are cached to JSON and can be resumed with `-resume`
- **Graceful Ctrl+C handling** — Stops all workers, offers to export partial results before exit
- **Large CIDR block protection** — Prevents memory exhaustion by limiting to 1M IPs per block

### Output
- **Text and JSON formats** — Structured JSON output with backward-compatible field names
- **Auto-export** — Results saved to file with `--export` or prompted after scan
- **Custom output directory** — Export files to a specific directory with `-output-dir`
- **Real-time progress bar** — Visual scan progress with ETA

### Configuration
- **YAML config file** — Set defaults in a config file, CLI flags always override
- **Auto-discovery** — Automatically finds config files in common locations
- **Proxy support** — HTTP, HTTPS, and SOCKS5 proxies with authentication
- **Custom DNS servers** — Use your own DNS resolvers for all lookups
- **Input validation** — Validates ASN, IP/CIDR, and domain formats before scanning

## Installation

**From Releases:**
```bash
# Download from releases
unzip ipmap.zip
chmod +x ipmap
./ipmap
```

**Build from Source:**
```bash
git clone https://github.com/sercanarga/ipmap.git
cd ipmap
go mod tidy
go build -o ipmap .
```

**Cross-platform Build Scripts:**
```bash
# Linux/macOS
./build.sh

# Windows (PowerShell)
.\build.ps1
```

## Usage

### Parameters

```bash
-asn AS13335                         # Scan all IP blocks in the ASN
-ip 103.21.244.0/22                  # Scan specified IP blocks
-d example.com                       # Search for specific domain
-t 2000                              # Request timeout in ms (auto if not set)
--export                             # Auto-export results
-format json                         # Output format (text or json)
-workers 100                         # Concurrent workers (default: 100)
-v                                   # Verbose mode
-c                                   # Continue scanning even after domain match
-proxy http://127.0.0.1:8080         # Proxy URL (HTTP/HTTPS/SOCKS5)
-rate 50                             # Rate limit (requests/sec, 0 = unlimited)
-dns 8.8.8.8,1.1.1.1                 # Custom DNS servers
-ipv6                                # Enable IPv6 scanning
-config config.yaml                  # Load config from YAML file
-resume cache.json                   # Resume interrupted scan from cache
-output-dir ./exports                # Directory for export files
-insecure=false                      # Skip TLS certificate verification (default: true)
```

### Examples

```bash
# Basic ASN scan
ipmap -asn AS13335

# Find domain in ASN
ipmap -asn AS13335 -d example.com

# Scan IP blocks
ipmap -ip 103.21.244.0/22,103.22.200.0/22

# High-performance scan
ipmap -asn AS13335 -workers 200 -v

# With proxy and rate limiting
ipmap -asn AS13335 -proxy socks5://127.0.0.1:9050 -rate 50

# Resume an interrupted scan
ipmap -resume ipmap_AS13335_cache.json

# Export results to a specific directory in JSON format
ipmap -asn AS13335 -format json --export -output-dir ./results

# Full configuration
ipmap -asn AS13335 -d example.com -proxy http://127.0.0.1:8080 -rate 100 -workers 50 -dns 8.8.8.8 -v --export
```

### Configuration File

ipmap supports YAML configuration files. Create a `config.yaml` in your working directory or `~/.ipmap.yaml` for global defaults:

```yaml
# config.yaml
workers: 150
timeout: 3000
rate_limit: 50
proxy: "socks5://127.0.0.1:9050"
dns_servers:
  - "8.8.8.8"
  - "1.1.1.1"
ipv6: false
verbose: false
format: "text"
```

**Auto-discovery order:** The tool automatically looks for config files in this order:
1. `config.yaml` / `config.yml` (current directory)
2. `.ipmap.yaml` / `.ipmap.yml` (current directory)
3. `~/.ipmap.yaml` / `~/.ipmap.yml` (home directory)

> **Note:** CLI flags always override config file values. Config file values only apply if the corresponding flag is not explicitly set on the command line.

### Scan Resume (Cache)

When a scan is interrupted (Ctrl+C), ipmap automatically offers to export partial results. You can also resume from where you left off:

```bash
# Start a scan (press Ctrl+C to interrupt)
ipmap -asn AS13335

# Resume the interrupted scan
ipmap -resume ipmap_AS13335_cache.json
```

The cache file (JSON) stores: scanned IPs, results found so far, scan metadata, and progress. On resume, already-scanned IPs are skipped automatically.

### Output Examples

**Text format (default):**
```
==================== RESULT ====================
Method:        Search All ASN/IP
Search Site:   Example Site
Timeout:       2000ms
IP Blocks:     103.21.244.0/22,103.22.200.0/22
Found Websites:
  200 | https://103.21.244.5 | Example Site [host.example.com.]
  200 | https://103.21.244.12 | Another Site
  403 | https://103.22.200.1 | Cloudflare
================================================
```

**JSON format (`-format json`):**
```json
{
  "method": "Search All ASN/IP",
  "search_site": "Example Site",
  "timeout_ms": 2000,
  "ip_blocks": ["103.21.244.0/22", "103.22.200.0/22"],
  "found_websites": [
    ["200", "https://103.21.244.5", "Example Site", "host.example.com."],
    ["200", "https://103.21.244.12", "Another Site"],
    ["403", "https://103.22.200.1", "Cloudflare"]
  ],
  "timestamp": "2026-03-11T02:00:00+03:00"
}
```

### How It Works

1. **ASN Lookup** — Queries [RIPE Stat API](https://stat.ripe.net/) for all announced IP prefixes of the given ASN. Falls back to [RADB](https://www.radb.net/) if RIPE is unavailable.
2. **IP Expansion** — Converts CIDR blocks to individual IP addresses (excluding network and broadcast addresses).
3. **IP Shuffling** — Randomizes the scan order to avoid sequential patterns that may trigger WAF/firewall rules.
4. **Parallel Scanning** — Distributes IPs across a configurable worker pool. Each worker:
   - Waits for the rate limiter (token bucket)
   - Adds random jitter (0-200ms)
   - Probes HTTPS first, falls back to HTTP
   - Extracts the `<title>` tag from the response
5. **Batch DNS** — After scanning, performs parallel reverse DNS lookups for all discovered IPs to resolve hostnames.
6. **Results** — Displays results with progress bar, prints summary, and offers export to file.

### Anti-Detection Details

ipmap uses multiple layers to avoid detection by WAFs (Cloudflare, Akamai, etc.):

| Layer | Technique |
|-------|-----------|
| **TLS** | Chrome 135 JA3/JA4 fingerprint via [uTLS](https://github.com/refraction-networking/utls) |
| **Headers** | Exact Chrome 135 header order (Host → sec-ch-ua → User-Agent → Accept → Sec-Fetch → etc.) |
| **User-Agent** | 17 different Chrome 133-135 variants (Windows/macOS/Linux) |
| **Referer** | Random rotation: Google, Bing, DuckDuckGo, Yahoo |
| **Scan Order** | IP addresses shuffled to prevent sequential detection |
| **Timing** | 0-200ms random jitter between requests |
| **Rate** | Configurable token bucket rate limiter |
| **Proxy** | HTTP/HTTPS/SOCKS5 proxy with authentication support |

## License

This project is open-source and available under the MIT License.

## Contributors
Thanks go to these wonderful people
<table>
  <tbody>
    <tr>
      <td align="center">
        <a href="https://github.com/ertugrulturan">
          <img src="https://avatars.githubusercontent.com/u/189706154?v=4" width="100px;" alt=""/>
          <br />
          <sub>
            <b>Ertuğrul TURAN</b>
          </sub>
        </a>
      </td>
      <td align="center">
        <a href="https://github.com/sametcodes">
          <img src="https://avatars.githubusercontent.com/u/9467273?v=4" width="100px;" alt=""/>
          <br />
          <sub>
            <b>Samet</b>
          </sub>
        </a>
      </td>
      <td align="center">
        <a href="https://github.com/lordixir">
          <img src="https://avatars.githubusercontent.com/u/38049901?v=4" width="100px;" alt=""/>
          <br />
          <sub>
            <b>Murat</b>
          </sub>
        </a>
      </td>
    </tr>
  </tbody>
</table>
