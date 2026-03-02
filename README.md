# ipmap

An open-source, cross-platform powerful network analysis tool for discovering websites hosted on specific IP addresses and ASN ranges.

## Features

- ASN scanning (Autonomous System Number) with IPv4/IPv6 support
- IP block scanning (CIDR format)
- HTTPS/HTTP automatic fallback
- **Chrome 135 TLS Fingerprint** (JA3/JA4 spoofing via uTLS)
- **Real Chrome Header Order** (WAF bypass optimized)
- **Referer Header Rotation** (Google, Bing, DuckDuckGo)
- Firewall bypass techniques (IP shuffling, header randomization, smart jitter)
- Proxy support (HTTP/HTTPS/SOCKS5)
- Custom DNS servers
- Rate limiting (token bucket algorithm)
- Dynamic timeout calculation
- Text and JSON output formats
- Configurable concurrent workers (1-1000)
- Real-time progress bar
- Graceful Ctrl+C handling with result export
- Input validation (ASN, IP/CIDR format checking)
- Large CIDR block protection (max 1M IPs)

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
-c                                   # Continue until completion
-proxy http://127.0.0.1:8080         # Proxy URL (HTTP/HTTPS/SOCKS5)
-rate 50                             # Rate limit (requests/sec, 0 = unlimited)
-dns 8.8.8.8,1.1.1.1                 # Custom DNS servers
-ipv6                                # Enable IPv6 scanning
-config config.yaml                  # Load config from YAML file
-resume cache.json                   # Resume interrupted scan from cache
-output-dir ./exports                # Directory for export files
-insecure=false                      # Enable TLS certificate verification
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

# Full configuration
ipmap -asn AS13335 -d example.com -proxy http://127.0.0.1:8080 -rate 100 -workers 50 -dns 8.8.8.8 -v --export
```

## License

This project is open-source and available under the MIT License.

## Contributors
Thanks go to these wonderful people
<table>
  <tbody>
    <tr>
      <td align="center">
        <a href="https://github.com/ertugrulturan">
          <img src="" width="100px;" alt=""/>
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
