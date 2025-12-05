# ipmap

An open-source, cross-platform powerful network analysis tool for discovering websites hosted on specific IP addresses and ASN ranges.

## Features
- ASN scanning (Autonomous System Number)
- IP block scanning (CIDR format)
- HTTPS/HTTP support
- DNS resolution
- Text and JSON output formats
- Configurable concurrent workers (1-1000)
- Real-time progress bar
- Graceful interrupt handling with result export

## Installation

Download the latest version from [Releases](https://github.com/sercanarga/ipmap/releases) and run:

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
-t 200                               # Request timeout in milliseconds
--export                             # Auto-export results
-format json                         # Output format (text or json)
-workers 100                         # Number of concurrent workers
-v                                   # Verbose mode
-c                                   # Continue scanning until completion
```

### Examples

**Scan ASN:**
```bash
ipmap -asn AS13335 -t 300
```

**Find domain in ASN:**
```bash
ipmap -asn AS13335 -d example.com
```

**Scan IP blocks:**
```bash
ipmap -ip 103.21.244.0/22,103.22.200.0/22 -t 300
```

**Export results:**
```bash
ipmap -asn AS13335 -d example.com --export
```

**High-performance scan:**
```bash
ipmap -asn AS13335 -workers 200 -v
```

## Proxy Usage

ipmap supports HTTP, HTTPS, and SOCKS5 proxies for anonymous scanning and bypassing network restrictions.

### Proxy Parameters
```bash
-proxy http://127.0.0.1:8080        # HTTP proxy
-proxy https://127.0.0.1:8080       # HTTPS proxy
-proxy socks5://127.0.0.1:1080      # SOCKS5 proxy
-rate 50                            # Rate limit (requests/second)
-dns 8.8.8.8,1.1.1.1               # Custom DNS servers
```

### Proxy Examples

**Basic HTTP proxy:**
```bash
ipmap -asn AS13335 -proxy http://127.0.0.1:8080
```

**SOCKS5 proxy with Tor:**
```bash
ipmap -asn AS13335 -proxy socks5://127.0.0.1:9050
```

**Proxy with authentication:**
```bash
ipmap -asn AS13335 -proxy http://user:password@proxy.example.com:8080
```

**Proxy with rate limiting:**
```bash
ipmap -asn AS13335 -proxy http://127.0.0.1:8080 -rate 50
```

**Proxy with custom DNS:**
```bash
ipmap -asn AS13335 -proxy socks5://127.0.0.1:1080 -dns 8.8.8.8,1.1.1.1
```

**Full configuration example:**
```bash
ipmap -asn AS13335 -d example.com -proxy http://127.0.0.1:8080 -rate 100 -workers 50 -v --export
```

> **Note:** When using proxies, consider reducing the worker count (`-workers`) and enabling rate limiting (`-rate`) to avoid overwhelming the proxy server.

## Building

```bash
git clone https://github.com/sercanarga/ipmap.git
cd ipmap
go build -o ipmap .
```

## Testing

```bash
go test ./... -v
```

## License

This project is open-source and available under the MIT License.
