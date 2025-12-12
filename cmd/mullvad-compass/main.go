// Package main provides the command-line interface for mullvad-compass.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/api"
	"github.com/Ch00k/mullvad-compass/internal/cli"
	"github.com/Ch00k/mullvad-compass/internal/formatter"
	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/ping"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

var Version = "dev"

// Dependencies encapsulates external dependencies for testing
type Dependencies struct {
	GetUserLocation func(context.Context, logging.LogLevel) (*api.UserLocation, error)
	PingLocations   func(context.Context, []relays.Location, int, int, relays.IPVersion, logging.LogLevel) ([]relays.Location, error)
	ParseRelaysFile func(logging.LogLevel, string, func() (string, error)) (*relays.File, error)
	Stdout          io.Writer
}

// DefaultDependencies returns production dependencies
func DefaultDependencies() Dependencies {
	return Dependencies{
		GetUserLocation: makeGetUserLocation(Version),
		PingLocations:   makePingLocations(),
		ParseRelaysFile: parseRelaysFile,
		Stdout:          os.Stdout,
	}
}

// makePingLocations creates a PingLocations function that accepts logLevel
func makePingLocations() func(context.Context, []relays.Location, int, int, relays.IPVersion, logging.LogLevel) ([]relays.Location, error) {
	return func(ctx context.Context, locations []relays.Location, timeout, workers int, ipVersion relays.IPVersion, logLevel logging.LogLevel) ([]relays.Location, error) {
		return ping.LocationsWithFactory(
			ctx,
			locations,
			timeout,
			workers,
			ipVersion,
			ping.NewDefaultPingerFactory(),
			logLevel,
		)
	}
}

// makeGetUserLocation creates a GetUserLocation function with the given version
func makeGetUserLocation(version string) func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
	return func(ctx context.Context, logLevel logging.LogLevel) (*api.UserLocation, error) {
		client := api.NewClient(api.WithVersion(version), api.WithLogLevel(logLevel))
		return client.GetUserLocation(ctx)
	}
}

func main() {
	// Create a context that can be cancelled with SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	if err := run(ctx, os.Args[1:], DefaultDependencies()); err != nil {
		// Don't print error if user cancelled with Ctrl-C
		if err == context.Canceled {
			fmt.Fprintln(os.Stderr, "Operation cancelled")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		cancel()
		os.Exit(1)
	}
	cancel()
}

// runBestServerMode finds and returns the best server by progressively expanding search radius
func runBestServerMode(
	ctx context.Context,
	logLevel logging.LogLevel,
	locations []relays.Location,
	userLoc *api.UserLocation,
	timeout int,
	workers int,
	ipVersion relays.IPVersion,
	stdout io.Writer,
	pingFn func(context.Context, []relays.Location, int, int, relays.IPVersion, logging.LogLevel) ([]relays.Location, error),
	deterministicOutput bool,
) error {
	currentRange := 500.0
	maxRange := 20000.0
	var filteredLocations []relays.Location

	for len(filteredLocations) == 0 {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}

		filteredLocations = filterByDistance(logLevel, locations, userLoc.Latitude, userLoc.Longitude, currentRange)
		if len(filteredLocations) == 0 {
			currentRange += 500.0
			if currentRange > maxRange {
				return fmt.Errorf("no servers found within maximum search radius of %.0f km", maxRange)
			}
		}
	}

	// Ping all servers in the found range
	var err error
	filteredLocations, err = pingLocations(
		ctx,
		logLevel,
		filteredLocations,
		timeout,
		workers,
		ipVersion,
		pingFn,
	)
	if err != nil {
		return err
	}

	// Sort by latency and return only the best server
	if len(filteredLocations) > 0 {
		sortLocationsByLatency(logLevel, filteredLocations)

		// Replace with deterministic data if flag is set
		bestServer := filteredLocations[0]
		deterministicUserLoc := userLoc
		if deterministicOutput {
			deterministicLocations := getDeterministicLocations()
			if len(deterministicLocations) > 0 {
				bestServer = deterministicLocations[0]
			}
			detUserLoc := getDeterministicUserLocation()
			deterministicUserLoc = &detUserLoc
		}

		output := formatter.FormatBestServer(*deterministicUserLoc, bestServer, ipVersion.IsIPv6())
		_, _ = fmt.Fprint(stdout, output)
	}

	return nil
}

func run(ctx context.Context, args []string, deps Dependencies) error {
	// Parse command-line flags
	config, err := cli.ParseFlags(args, Version)
	if err != nil {
		return err
	}
	if config.LogLevel <= logging.LogLevelDebug {
		log.Printf("Config: %+v", config)
	}

	// Handle help flag
	if config.ShowHelp {
		cli.PrintUsage(deps.Stdout, Version)
		return nil
	}

	// Handle version flag
	if config.ShowVersion {
		_, _ = fmt.Fprintf(deps.Stdout, "mullvad-compass %s\n", Version)
		return nil
	}

	// Start timing for the entire operation
	operationStart := time.Now()
	defer func() {
		if config.LogLevel <= logging.LogLevelDebug {
			elapsed := time.Since(operationStart)
			log.Printf("Total operation completed in %v", elapsed)
		}
	}()

	// Parse relays file
	if config.LogLevel <= logging.LogLevelDebug {
		log.Println("Parsing relays file...")
	}
	relaysData, err := deps.ParseRelaysFile(config.LogLevel, "", relays.GetRelaysFilePath)
	if err != nil {
		return err
	}

	// Get locations from relays file, optionally filtered by anti-censorship, DAITA, and IPv6
	if config.LogLevel <= logging.LogLevelDebug {
		log.Println("Fetching and filtering relay locations...")
	}
	locations, err := getLocations(
		config.LogLevel,
		relaysData,
		config.AntiCensorship,
		config.Daita,
		config.IPVersion,
	)
	if err == nil {
		if config.LogLevel <= logging.LogLevelDebug {
			log.Printf("Found %d matching servers", len(locations))
		}
	} else {
		return err
	}

	if len(locations) == 0 {
		return fmt.Errorf("no servers found")
	}

	// Get user location
	if config.LogLevel <= logging.LogLevelDebug {
		log.Println("Fetching user location...")
	}
	userLoc, err := getUserLocation(ctx, config.LogLevel, deps.GetUserLocation)
	if err != nil {
		return fmt.Errorf("failed to get user location: %w", err)
	}

	// Check if connected to Mullvad VPN
	if userLoc.MullvadExitIP {
		userLocOutput := formatter.FormatUserLocation(*userLoc)
		_, _ = fmt.Fprintf(
			deps.Stdout,
			"You are currently connected to Mullvad VPN. Pinging Mullvad servers from a Mullvad server does not provide meaningful results.\nYour location info:\n%s",
			userLocOutput,
		)
		return nil
	}

	// Best server mode: progressively expand range until we find servers
	if config.BestServerMode {
		return runBestServerMode(
			ctx,
			config.LogLevel,
			locations,
			userLoc,
			config.Timeout,
			config.Workers,
			config.IPVersion,
			deps.Stdout,
			deps.PingLocations,
			config.DeterministicOutput,
		)
	}

	// Normal mode: filter by distance
	if config.LogLevel <= logging.LogLevelDebug {
		log.Printf("Filtering servers within %.0f km...", config.MaxDistance)
	}
	locations = filterByDistance(config.LogLevel, locations, userLoc.Latitude, userLoc.Longitude, config.MaxDistance)

	if config.LogLevel <= logging.LogLevelDebug {
		serverWord := "servers"
		if len(locations) == 1 {
			serverWord = "server"
		}
		log.Printf("%d %s found within %.0f km", len(locations), serverWord, config.MaxDistance)
	}

	if len(locations) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No servers found within %.0f km of your location\n", config.MaxDistance)
		return nil
	}

	// Ping locations
	if config.LogLevel <= logging.LogLevelDebug {
		log.Println("Pinging servers...")
	}
	locations, err = pingLocations(
		ctx,
		config.LogLevel,
		locations,
		config.Timeout,
		config.Workers,
		config.IPVersion,
		deps.PingLocations,
	)
	if err != nil {
		return err
	}

	// Sort and display results
	if config.LogLevel <= logging.LogLevelDebug {
		log.Println("Sorting servers by latency...")
	}
	sortLocationsByLatency(config.LogLevel, locations)

	// Replace with deterministic data if flag is set
	if config.DeterministicOutput {
		locations = getDeterministicLocations()
	}

	table := formatter.FormatTable(locations, config.IPVersion.IsIPv6())
	_, _ = fmt.Fprint(deps.Stdout, table)

	return nil
}
