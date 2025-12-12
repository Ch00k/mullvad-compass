# mullvad-compass

Find Mullvad VPN servers with the lowest latency at your current geographic location.

A rewrite of [mullvad-closest](https://github.com/Ch00k/mullvad-closest), offering more features, faster performance, and a single binary distribution with no runtime dependencies.

## Features

- Finds the single best or multiple closest Mullvad VPN servers
- Filters by distance threshold and anti-censorship protocol
- Measures actual latency via ICMP ping
- Supports IPv4 and IPv6 addresses
- Executes concurrent pings
- Shows your current location based on Mullvad's API
- Platform support: Linux, macOS, and Windows

## Installation

Download pre-built binaries from the [releases page](https://github.com/Ch00k/mullvad-compass/releases).

### Mullvad VPN

The tool reads Mullvad's `relays.json` file from the platform-specific location:

- **Linux**: `/var/cache/mullvad-vpn/relays.json`
- **macOS**: `/Library/Caches/mullvad-vpn/relays.json`
- **Windows**: `C:/ProgramData/Mullvad VPN/cache/relays.json`

This file is created when you install the Mullvad VPN client.

## Usage

Run without options to find the single best (lowest latency) server:

<!-- best-server:start -->
```
$ mullvad-compass
Your location:   Dresden, Germany
                 203.0.113.42
Best server:     Prague, Czech Republic
                 cz-prg-wg-201 (178.249.209.162)
                 9.78 ms, 156 km away
```
<!-- best-server:end -->

Find multiple servers within 200 km, sorted by latency (lowest first):

<!-- multiple-servers:start -->
```
$ mullvad-compass --max-distance 250
Country          City     Distance (km)   Hostname        IP                Latency (ms)
--------------   ------   -------------   -------------   ---------------   ------------
Czech Republic   Prague   156             cz-prg-wg-201   178.249.209.162   9.78
Czech Republic   Prague   156             cz-prg-wg-202   178.249.209.175   13.01
Czech Republic   Prague   156             cz-prg-wg-102   146.70.129.130    13.94
Germany          Berlin   238             de-ber-wg-007   193.32.248.75     15.86
Germany          Berlin   238             de-ber-wg-001   193.32.248.66     15.88
Germany          Berlin   238             de-ber-wg-005   193.32.248.70     15.89
Germany          Berlin   238             de-ber-wg-008   193.32.248.74     15.91
Germany          Berlin   238             de-ber-wg-003   193.32.248.68     15.93
Germany          Berlin   238             de-ber-wg-004   193.32.248.69     15.95
Germany          Berlin   238             de-ber-wg-006   193.32.248.71     15.95
Germany          Berlin   238             de-ber-wg-002   193.32.248.67     15.99
```
<!-- multiple-servers:end -->

All options can be viewed with `--help`:

<!-- help:start -->
```
$ mullvad-compass --help
mullvad-compass 0.0.2

Find Mullvad VPN servers with the lowest latency at your current location.

USAGE:
    mullvad-compass [OPTIONS]

MODES:
    Best Server Mode (default):   Shows your location and the single best server.
                                  Activated when running without filter options.

    Table Mode:                   Shows all matching servers in a table, sorted by latency.
                                  Activated by using any filter option (-m, -a, -d, -6).

FILTER OPTIONS (Table Mode):
    -m, --max-distance KM         Maximum distance in km from your location (default: 500, range: 1-20000)
    -a, --anti-censorship TYPE    Filter servers by anti-censorship type (lwo, quic, shadowsocks)
    -d, --daita                   Filter servers with DAITA enabled
    -6, --ipv6                    Use IPv6 addresses for pinging

PERFORMANCE OPTIONS:
    -t, --timeout MS              Ping timeout in milliseconds (default: 500, range: 100-5000)
    -w, --workers COUNT           Number of concurrent ping workers (default: 25, range: 1-200)

OTHER OPTIONS:
    -l, --log-level LEVEL         Set log level (debug, info, warning, error; default: error)
    -h, --help                    Show this help message
    -v, --version                 Show version information
```
<!-- help:end -->

## How It Works

1. Fetches your current location from Mullvad's API (`https://am.i.mullvad.net/json`)
2. Reads the Mullvad server list from `relays.json`
3. Calculates geodesic distances using the Haversine formula
4. Filters servers within the specified distance threshold, and based filters specified
5. Pings servers concurrently
6. Displays results sorted by latency (lowest first)
