# ipmap

An open-source, cross-platform powerful network analysis tool for discovering websites hosted on specific IP addresses and ASN ranges.

## Features

- ASN scanning (Autonomous System Number) with IPv4/IPv6 support
- IP block scanning (CIDR format)
- HTTPS/HTTP automatic fallback
- Firewall bypass techniques (IP shuffling, header randomization, jitter)
- Proxy support (HTTP/HTTPS/SOCKS5)
- Custom DNS servers
- Rate limiting (token bucket algorithm)
- Dynamic timeout calculation
- Text and JSON output formats
- Configurable concurrent workers (1-1000)
- Real-time progress bar
- Graceful Ctrl+C handling with result export

## Installation

Download the latest version from [Releases](https://github.com/lordixir/ipmap/releases) and run:

```bash
unzip ipmap.zip
chmod +x ipmap
./ipmap
```

## Usage

### Parameters

```bash
-asn AS13335                         # Scan all IP blocks in the ASN
-ip 103.21.244.0/22                  # Scan specified IP blocks
-d example.com                       # Search for specific domain
-t 2000                              # Request timeout in milliseconds (auto-calculated if not set)
--export                             # Auto-export results
-format json                         # Output format (text or json)
-workers 100                         # Number of concurrent workers (default: 100)
-v                                   # Verbose mode
-c                                   # Continue scanning until completion
-proxy http://127.0.0.1:8080         # Proxy URL (HTTP/HTTPS/SOCKS5)
-rate 50                             # Rate limit (requests/second, 0 = unlimited)
-dns 8.8.8.8,1.1.1.1                # Custom DNS servers
```

### Examples

**Basic ASN scan (auto timeout):**
```bash
ipmap -asn AS13335
```

**Find domain in ASN:**
```bash
ipmap -asn AS13335 -d example.com
```

**Scan IP blocks:**
```bash
ipmap -ip 103.21.244.0/22,103.22.200.0/22
```

**High-performance scan:**
```bash
ipmap -asn AS13335 -workers 200 -v
```

**Export results:**
```bash
ipmap -asn AS13335 -d example.com --export
```

**JSON output:**
```bash
ipmap -asn AS13335 -format json --export
```

## Proxy & Rate Limiting

ipmap supports HTTP, HTTPS, and SOCKS5 proxies for anonymous scanning.

**HTTP proxy:**
```bash
ipmap -asn AS13335 -proxy http://127.0.0.1:8080
```

**SOCKS5 proxy (Tor):**
```bash
ipmap -asn AS13335 -proxy socks5://127.0.0.1:9050
```

**Proxy with auth:**
```bash
ipmap -asn AS13335 -proxy http://user:pass@proxy.com:8080
```

**Rate limiting:**
```bash
ipmap -asn AS13335 -rate 50 -workers 50
```

**Full configuration:**
```bash
ipmap -asn AS13335 -d example.com -proxy http://127.0.0.1:8080 -rate 100 -workers 50 -dns 8.8.8.8 -v --export
```

> **Note:** When using proxies, reduce worker count and enable rate limiting to avoid overwhelming the proxy.

## Firewall Bypass Features

ipmap includes built-in firewall bypass techniques:

- **IP Shuffling:** Randomizes scan order to avoid sequential pattern detection
- **Header Randomization:** Rotates User-Agent, Accept-Language, Chrome versions, platforms
- **Request Jitter:** Adds random 0-50ms delay between requests
- **Dynamic Timeout:** Auto-adjusts timeout based on worker count

## Interrupt Handling (Ctrl+C)

Press Ctrl+C during scan to:
1. Immediately stop all scanning
2. View found results count
3. Option to export partial results

## Building

```bash
git clone https://github.com/lordixir/ipmap.git
cd ipmap
go build -o ipmap .
```

## Testing

```bash
go test ./... -v
```

## Changelog (v2.0)

- ✅ Added IP shuffling for firewall bypass
- ✅ Added request jitter (0-50ms random delay)
- ✅ Added header randomization (language, chrome version, platform)
- ✅ Fixed Ctrl+C interrupt handling (immediate stop)
- ✅ Added dynamic timeout calculation based on workers
- ✅ Added IPv6 support for ASN scanning
- ✅ Improved error logging
- ✅ Fixed result collection bug with high workers
- ✅ Removed gzip to fix response parsing
- ✅ Added scan statistics at completion

## License

This project is open-source and available under the MIT License.

## Contributors
Thanks go to these wonderful people
<table>
  <tbody>
    <tr>
      <td align="center">
        <a href="https://github.com/ertugrulturan">
          <img src="https://avatars.githubusercontent.com/u/60829297?v=4" width="100px;" alt=""/>
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


