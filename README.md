# mullvad-compass

Find Mullvad VPN servers with the lowest latency at your current geographic location.

A Go rewrite of [mullvad-closest](https://github.com/Ch00k/mullvad-closest), offering faster performance and a single binary distribution with no runtime dependencies.

## Features

- Finds Mullvad VPN servers within a specified distance from your location
- Measures actual latency via ICMP ping
- Filters by server type (OpenVPN or WireGuard)
- Concurrent pinging for fast results
- Platform support: Linux, macOS, and Windows

## Installation

### From Source

```bash
go install github.com/Ch00k/mullvad-compass@latest
```

Or clone and build:

```bash
git clone https://github.com/Ch00k/mullvad-compass
cd mullvad-compass
go build -o mullvad-compass
```

### Binary Releases

Download pre-built binaries from the [releases page](https://github.com/Ch00k/mullvad-compass/releases).

## Requirements

### ICMP Privileges

This tool requires the ability to send ICMP packets (raw sockets). You have two options:

#### Linux

Grant the binary `CAP_NET_RAW` capability:

```bash
sudo setcap cap_net_raw+ep /path/to/mullvad-compass
```

Or run as root:

```bash
sudo ./mullvad-compass
```

#### macOS

Run as root:

```bash
sudo ./mullvad-compass
```

#### Windows

Run as Administrator.

### Mullvad VPN

The tool reads Mullvad's `relays.json` file from the platform-specific location:

- **Linux**: `/var/cache/mullvad-vpn/relays.json`
- **macOS**: `/Library/Caches/mullvad-vpn/relays.json`
- **Windows**: `C:/ProgramData/Mullvad VPN/cache/relays.json`

This file is created when you install the Mullvad VPN client. If you don't have Mullvad VPN installed, you can download the relays.json file and specify its path with `--relays-file`.

## Usage

### Basic Usage

Find all servers within 500 km (default):

```bash
./mullvad-compass
```

### Filter by Server Type

Show only WireGuard servers:

```bash
./mullvad-compass -s wireguard
```

Show only OpenVPN servers:

```bash
./mullvad-compass -s openvpn
```

### Custom Distance Threshold

Find servers within 300 km:

```bash
./mullvad-compass -m 300
```

### Custom Relays File

Specify a custom path to relays.json:

```bash
./mullvad-compass --relays-file /path/to/relays.json
```

### Combined Options

```bash
./mullvad-compass -s wireguard -m 1000
```

## Output

Results are displayed in a formatted table sorted by latency:

```
Country  City      Type       IP              Hostname           Distance  Latency
-------  --------  ---------  --------------  -----------------  --------  -------
Germany  Frankfurt wireguard  185.65.134.130  de-fra-wg-301      245.67    12.34
Germany  Berlin    wireguard  185.65.134.145  de-ber-wg-201      234.89    15.67
Poland   Warsaw    wireguard  146.70.166.2    pl-waw-wg-001      432.11    23.45
```

- **Distance**: Geodesic distance in kilometers (Haversine formula)
- **Latency**: Round-trip ICMP ping time in milliseconds
- Servers that timeout or are unreachable show "timeout"

## How It Works

1. Fetches your current location from Mullvad's API (`https://am.i.mullvad.net/json`)
2. Reads the Mullvad server list from `relays.json`
3. Calculates geodesic distances using the Haversine formula
4. Filters servers within the specified distance threshold
5. Pings servers concurrently (up to 25 workers)
6. Displays results sorted by latency

## Development

### Running Tests

```bash
go test -v
```

### Building

```bash
go build -o mullvad-compass
```

### Cross-Compilation

For Linux:
```bash
GOOS=linux GOARCH=amd64 go build -o mullvad-compass-linux-amd64
```

For macOS:
```bash
GOOS=darwin GOARCH=amd64 go build -o mullvad-compass-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o mullvad-compass-darwin-arm64
```

For Windows:
```bash
GOOS=windows GOARCH=amd64 go build -o mullvad-compass-windows-amd64.exe
```

## Architecture

The codebase is organized into simple, focused files:

- `main.go` - CLI entry point and flag parsing
- `location.go` - Location data type
- `api.go` - Mullvad API client
- `relays.go` - Relays.json parser with platform-specific paths
- `distance.go` - Haversine distance calculation
- `ping.go` - ICMP ping implementation with concurrency
- `formatter.go` - Table output formatting
- `testdata/` - Test fixtures (sample relays.json)

## Dependencies

- Go standard library
- `golang.org/x/net/icmp` - ICMP protocol implementation
- `golang.org/x/net/ipv4` - IPv4 packet handling

The `golang.org/x/net` packages are maintained by the Go team and provide low-level networking functionality not available in the standard library.

## Comparison with Python Version

### Advantages

- **Faster execution**: Compiled binary vs interpreted Python
- **Better concurrency**: Goroutines vs ThreadPoolExecutor
- **Single binary**: No Python runtime or dependencies required
- **Smaller footprint**: ~8.5 MB binary vs Python + dependencies

### Differences

- **Distance calculation**: Uses Haversine formula instead of Vincenty
  - For distances under 500 km, the difference is negligible (<0.5% error)
  - Haversine is simpler and doesn't require external dependencies
- **Binary name**: `mullvad-compass` instead of `mullvad-closest`

## License

[Unlicense](LICENSE) (public domain)

## Author

Andrii Yurchuk (ay@mntw.re)

## Contributing

Issues and pull requests are welcome at https://github.com/Ch00k/mullvad-compass
