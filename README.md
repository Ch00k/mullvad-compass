# mullvad-compass

Find Mullvad VPN servers with the lowest latency at your current geographic location.

A rewrite of [mullvad-closest](https://github.com/Ch00k/mullvad-closest), offering more features, faster performance, and a single binary distribution with no runtime dependencies.

## Features

- Finds the single best or multiple closest Mullvad VPN servers
- Filters by server type (OpenVPN or WireGuard), distance threshold, WireGuard obfuscation type
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
Country:    Czech Republic
City:       Prague
Distance:   156 km
Hostname:   cz-prg-wg-201
IP:         178.249.209.162
Latency:    9.85 ms
```
<!-- best-server:end -->

Find multiple servers within 200 km, sorted by latency (lowest first):

<!-- multiple-servers:start -->
```
$ mullvad-compass --max-distance 250
Country          City     Type        IP                Hostname          Distance (km)   Latency (ms)
--------------   ------   ---------   ---------------   ---------------   -------------   ------------
Czech Republic   Prague   wireguard   178.249.209.162   cz-prg-wg-201     156             9.85
Czech Republic   Prague   openvpn     146.70.129.194    cz-prg-ovpn-102   156             12.86
Czech Republic   Prague   openvpn     146.70.129.162    cz-prg-ovpn-101   156             12.89
Czech Republic   Prague   wireguard   178.249.209.175   cz-prg-wg-202     156             12.95
Czech Republic   Prague   wireguard   146.70.129.130    cz-prg-wg-102     156             13.94
Germany          Berlin   wireguard   193.32.248.66     de-ber-wg-001     238             15.85
Germany          Berlin   wireguard   193.32.248.68     de-ber-wg-003     238             15.87
Germany          Berlin   wireguard   193.32.248.74     de-ber-wg-008     238             15.88
Germany          Berlin   wireguard   193.32.248.75     de-ber-wg-007     238             15.88
Germany          Berlin   wireguard   193.32.248.69     de-ber-wg-004     238             15.89
Germany          Berlin   wireguard   193.32.248.67     de-ber-wg-002     238             15.93
Germany          Berlin   openvpn     193.32.248.72     de-ber-ovpn-001   238             15.94
Germany          Berlin   wireguard   193.32.248.70     de-ber-wg-005     238             15.96
Germany          Berlin   wireguard   193.32.248.71     de-ber-wg-006     238             16.03
```
<!-- multiple-servers:end -->

All options can be viewed with `--help`:

<!-- help:start -->
```
$ mullvad-compass --help
mullvad-compass 0.0.1

Find Mullvad VPN servers with the lowest latency at your current location.

USAGE:
    mullvad-compass [OPTIONS]

Running without options finds the single best (closest, fastest, lowest latency) server among all available servers.

OPTIONS:
    -m, --max-distance KM              Maximum distance in km from your location (default: 500, range: 1-20000)
    -s, --server-type TYPE             Filter by server type (wireguard, openvpn)
    -o, --wireguard-obfuscation TYPE   Filter WireGuard servers by obfuscation (lwo, quic, shadowsocks)
    -d, --daita                        Filter WireGuard servers with DAITA enabled
    -6, --ipv6                         Use IPv6 addresses for pinging
    -t, --timeout MS                   Ping timeout in milliseconds (default: 500, range: 100-5000)
    -w, --workers COUNT                Number of concurrent ping workers (default: 25, range: 1-200)
    -i, --where-am-i                   Show your current location
    -l, --log-level LEVEL              Set log level (debug, info, warning, error; default: error)
    -h, --help                         Show this help message
    -v, --version                      Show version information
```
<!-- help:end -->

## How It Works

1. Fetches your current location from Mullvad's API (`https://am.i.mullvad.net/json`)
2. Reads the Mullvad server list from `relays.json`
3. Calculates geodesic distances using the Haversine formula
4. Filters servers within the specified distance threshold, and based filters specified
5. Pings servers concurrently
6. Displays results sorted by latency (lowest first)

## ICMP Ping Requirements

mullvad-compass measures server latency using ICMP echo (ping) packets. The privilege requirements vary by operating system.

### Linux

Linux supports two types of ICMP sockets:

**Option 1: Unprivileged UDP ICMP (Recommended)**

Modern Linux systems allow non-root users to send ICMP echo packets via UDP datagram sockets. This requires configuring the `ping_group_range` sysctl parameter:

```bash
# Check current setting
cat /proc/sys/net/ipv4/ping_group_range

# Allow all users (temporary, until reboot)
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"

# Make permanent
echo "net.ipv4.ping_group_range=0 2147483647" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

The default value varies by distribution:
- Ubuntu: Often `1 0` (disabled)
- Some distributions: `0 2147483647` (all users allowed)

**Option 2: Raw ICMP with Capabilities**

Grant the binary the `CAP_NET_RAW` capability:

```bash
sudo setcap cap_net_raw+ep /path/to/mullvad-compass
```

**Option 3: Run as Root**

```bash
sudo mullvad-compass
```

**Behavior**: The application automatically tries raw ICMP first, then falls back to UDP ICMP if that fails. If both fail, tests will error.

### macOS

macOS natively supports unprivileged UDP ICMP sockets without additional configuration. No special privileges or setup required.

### Windows

Windows requires raw ICMP sockets (`ip4:icmp` protocol). While the `golang.org/x/net/icmp` package uses privileged socket types on Windows, recent versions of Windows 10 and later typically allow this without administrator elevation.

**If you encounter permission errors:**

1. Run as Administrator, or
2. Check Windows Firewall settings

**Note**: Unlike Linux and macOS, Windows does not support unprivileged UDP-based ICMP in the Go networking stack.

### References

- [Unprivileged ICMP sockets on Linux](https://sturmflut.github.io/linux/ubuntu/2015/01/17/unprivileged-icmp-sockets-on-linux/)
- [golang.org/x/net/icmp package documentation](https://pkg.go.dev/golang.org/x/net/icmp)

## License

[Unlicense](LICENSE)
