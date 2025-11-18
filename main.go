package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const version = "0.0.1"

// Dependencies encapsulates external dependencies for testing
type Dependencies struct {
	GetUserLocation func() (*UserLocation, error)
	PingLocations   func([]Location) ([]Location, error)
	GetRelaysPath   func() (string, error)
	Stdout          io.Writer
}

// DefaultDependencies returns production dependencies
func DefaultDependencies() Dependencies {
	return Dependencies{
		GetUserLocation: GetUserLocation,
		PingLocations:   PingLocations,
		GetRelaysPath:   GetRelaysFilePath,
		Stdout:          os.Stdout,
	}
}

func main() {
	if err := run(os.Args[1:], DefaultDependencies()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, deps Dependencies) error {
	// Parse command-line flags
	config, err := parseFlags(args)
	if err != nil {
		return err
	}

	// Get relays file path (use provided path or platform-specific default)
	relaysPath := config.relaysPath
	if relaysPath == "" {
		defaultPath, err := deps.GetRelaysPath()
		if err != nil {
			return fmt.Errorf("could not find relays.json: %w\nPlease specify the path using --relays-file", err)
		}
		relaysPath = defaultPath
	}

	// Parse relays file
	relays, err := ParseRelaysFile(relaysPath)
	if err != nil {
		return err
	}

	// Get locations from relays file, optionally filtered by type
	locations, err := GetLocations(relays, config.serverType)
	if err != nil {
		return err
	}

	if len(locations) == 0 {
		return fmt.Errorf("no servers found")
	}

	// Get user location
	userLoc, err := deps.GetUserLocation()
	if err != nil {
		return fmt.Errorf("failed to get user location: %w", err)
	}

	// Filter by distance
	locations = FilterByDistance(locations, userLoc.Latitude, userLoc.Longitude, config.maxDistance)

	if len(locations) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No servers found within %.0f km of your location\n", config.maxDistance)
		return nil
	}

	// Ping locations
	locations, err = deps.PingLocations(locations)
	if err != nil {
		return err
	}

	// Format and display results
	table := FormatTable(locations)
	_, _ = fmt.Fprint(deps.Stdout, table)

	return nil
}

type config struct {
	serverType  string
	maxDistance float64
	relaysPath  string
}

// parseFlags parses command-line arguments manually to support GNU-style long flags
func parseFlags(args []string) (*config, error) {
	cfg := &config{
		maxDistance: 500.0,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			printUsage()
			os.Exit(0)

		case arg == "-v" || arg == "--version":
			fmt.Printf("mullvad-compass %s\n", version)
			os.Exit(0)

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

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			return nil, fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	return cfg, nil
}

func printUsage() {
	fmt.Printf(`mullvad-compass %s

Find Mullvad VPN servers with the lowest latency at your current location.

USAGE:
    mullvad-compass [OPTIONS]

OPTIONS:
    -s, --server-type TYPE      Filter by server type (openvpn or wireguard)
    -m, --max-distance DIST     Maximum distance in km from your location (default: 500)
    -r, --relays-file PATH      Path to relays.json file (auto-detected if not specified)
    -h, --help                  Show this help message
    -v, --version               Show version information

EXAMPLES:
    mullvad-compass
    mullvad-compass -s wireguard -m 300
    mullvad-compass --server-type openvpn --max-distance 1000
    mullvad-compass --relays-file /path/to/relays.json

NOTE:
    This tool requires CAP_NET_RAW capability or root privileges to send ICMP packets.
    On Linux, you can grant the capability with:
        sudo setcap cap_net_raw+ep /path/to/mullvad-compass
`, version)
}
