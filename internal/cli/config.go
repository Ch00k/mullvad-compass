// Package cli provides command-line interface configuration and flag parsing functionality.
package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// Config holds all command-line configuration options for the application.
type Config struct {
	AntiCensorship      relays.AntiCensorship
	Daita               bool
	IPVersion           relays.IPVersion
	MaxDistance         float64
	ShowHelp            bool
	ShowVersion         bool
	Timeout             int
	Workers             int
	BestServerMode      bool
	LogLevel            logging.LogLevel
	DeterministicOutput bool
}

// ParseFlags parses command-line arguments manually to support GNU-style long flags
func ParseFlags(args []string, version string) (*Config, error) {
	cfg := &Config{
		MaxDistance:    500.0,
		Timeout:        500,
		Workers:        25,
		BestServerMode: true,
		LogLevel:       logging.LogLevelError,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-h" || arg == "--help":
			cfg.ShowHelp = true
			return cfg, nil

		case arg == "-v" || arg == "--version":
			cfg.ShowVersion = true
			return cfg, nil

		case arg == "-a" || arg == "--anti-censorship":
			cfg.BestServerMode = false
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s requires an argument", arg)
			}
			i++
			antiCensorship, err := relays.ParseAntiCensorship(args[i])
			if err != nil {
				return nil, err
			}
			cfg.AntiCensorship = antiCensorship

		case arg == "-d" || arg == "--daita":
			cfg.BestServerMode = false
			cfg.Daita = true

		case arg == "-6" || arg == "--ipv6":
			cfg.BestServerMode = false
			cfg.IPVersion = relays.IPv6

		case arg == "-m" || arg == "--max-distance":
			cfg.BestServerMode = false
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
			if distance > 20000 {
				return nil, fmt.Errorf("max-distance must be at most 20000 km")
			}
			cfg.MaxDistance = distance

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
			cfg.Timeout = timeout

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
			cfg.Workers = workers

		case arg == "-l" || arg == "--log-level":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s requires an argument", arg)
			}
			i++
			level, err := logging.ParseLogLevel(args[i])
			if err != nil {
				return nil, err
			}
			cfg.LogLevel = level

		case arg == "--deterministic-output":
			// Only enable in dev builds, silently ignore otherwise
			if version == "dev" {
				cfg.DeterministicOutput = true
			}

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			return nil, fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	return cfg, nil
}

// PrintUsage outputs the usage information and command-line options to the writer.
func PrintUsage(w io.Writer, version string) {
	_, _ = fmt.Fprintf(w, `mullvad-compass %s

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
`, version)
}
