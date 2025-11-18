package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type config struct {
	serverType   string
	maxDistance  float64
	relaysPath   string
	showHelp     bool
	showVersion  bool
	showWhereAmI bool
	timeout      int
	workers      int
}

// parseFlags parses command-line arguments manually to support GNU-style long flags
func parseFlags(args []string) (*config, error) {
	cfg := &config{
		maxDistance: 500.0,
		timeout:     500,
		workers:     25,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			cfg.showHelp = true
			return cfg, nil

		case arg == "-v" || arg == "--version":
			cfg.showVersion = true
			return cfg, nil

		case arg == "-i" || arg == "--where-am-i":
			cfg.showWhereAmI = true
			return cfg, nil

		case arg == "-s" || arg == "--server-type":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s requires an argument", arg)
			}
			i++
			serverType := args[i]
			if serverType != "openvpn" && serverType != "wireguard" {
				return nil, fmt.Errorf("invalid server type: %s (must be 'openvpn' or 'wireguard')", serverType)
			}
			cfg.serverType = serverType

		case arg == "-m" || arg == "--max-distance":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s requires an argument", arg)
			}
			i++
			distance, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid max-distance value: %s", args[i])
			}
			if distance <= 0 {
				return nil, fmt.Errorf("max-distance must be positive")
			}
			cfg.maxDistance = distance

		case arg == "-r" || arg == "--relays-file":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s requires an argument", arg)
			}
			i++
			cfg.relaysPath = args[i]

		case arg == "-t" || arg == "--timeout":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s requires an argument", arg)
			}
			i++
			timeout, err := strconv.Atoi(args[i])
			if err != nil {
				return nil, fmt.Errorf("invalid timeout value: %s", args[i])
			}
			if timeout < 100 || timeout > 5000 {
				return nil, fmt.Errorf("timeout must be between 100 and 5000")
			}
			cfg.timeout = timeout

		case arg == "-w" || arg == "--workers":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s requires an argument", arg)
			}
			i++
			workers, err := strconv.Atoi(args[i])
			if err != nil {
				return nil, fmt.Errorf("invalid workers value: %s", args[i])
			}
			if workers < 1 || workers > 200 {
				return nil, fmt.Errorf("workers must be between 1 and 200")
			}
			cfg.workers = workers

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			return nil, fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	return cfg, nil
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintf(w, `mullvad-compass %s

Find Mullvad VPN servers with the lowest latency at your current location.

USAGE:
    mullvad-compass [OPTIONS]

OPTIONS:
    -s, --server-type TYPE      Filter by server type (openvpn or wireguard)
    -m, --max-distance DIST     Maximum distance in km from your location (default: 500)
    -r, --relays-file PATH      Path to relays.json file (auto-detected if not specified)
    -t, --timeout MS            Ping timeout in milliseconds (default: 500, range: 100-5000)
    -w, --workers COUNT         Number of concurrent ping workers (default: 25, range: 1-200)
    -i, --where-am-i            Display your current location and exit
    -h, --help                  Show this help message
    -v, --version               Show version information

EXAMPLES:
    mullvad-compass
    mullvad-compass -s wireguard -m 300
    mullvad-compass --server-type openvpn --max-distance 1000
    mullvad-compass --relays-file /path/to/relays.json
    mullvad-compass --where-am-i

NOTE:
    This tool requires CAP_NET_RAW capability or root privileges to send ICMP packets.
    On Linux, you can grant the capability with:
        sudo setcap cap_net_raw+ep /path/to/mullvad-compass
`, Version)
}
