package main

import (
	"fmt"
	"io"
	"os"
)

var Version = "0.0.1"

// Dependencies encapsulates external dependencies for testing
type Dependencies struct {
	GetUserLocation func() (*UserLocation, error)
	PingLocations   func([]Location, int, int) ([]Location, error)
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

	// Handle help flag
	if config.showHelp {
		printUsage(deps.Stdout)
		return nil
	}

	// Handle version flag
	if config.showVersion {
		_, _ = fmt.Fprintf(deps.Stdout, "mullvad-compass %s\n", Version)
		return nil
	}

	// Handle whereami flag
	if config.showWhereAmI {
		userLoc, err := deps.GetUserLocation()
		if err != nil {
			return fmt.Errorf("failed to get user location: %w", err)
		}
		mullvadStatus := "No"
		if userLoc.MullvadExitIP {
			mullvadStatus = "Yes"
		}
		_, _ = fmt.Fprintf(
			deps.Stdout,
			"IP: %s\nCity: %s\nCountry: %s\nLatitude: %f\nLongitude: %f\nConnected to Mullvad VPN: %s\n",
			userLoc.IP,
			userLoc.City,
			userLoc.Country,
			userLoc.Latitude,
			userLoc.Longitude,
			mullvadStatus,
		)
		return nil
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

	// Check if connected to Mullvad VPN
	if userLoc.MullvadExitIP {
		_, _ = fmt.Fprintf(
			deps.Stdout,
			"You are currently connected to Mullvad VPN. Pinging Mullvad servers from a Mullvad server does not provide meaningful results.\nDisconnect from the VPN and try again, or use --where-am-i to see your current location.\n",
		)
		return nil
	}

	// Filter by distance
	locations = FilterByDistance(locations, userLoc.Latitude, userLoc.Longitude, config.maxDistance)

	if len(locations) == 0 {
		_, _ = fmt.Fprintf(deps.Stdout, "No servers found within %.0f km of your location\n", config.maxDistance)
		return nil
	}

	// Ping locations
	locations, err = deps.PingLocations(locations, config.timeout, config.workers)
	if err != nil {
		return err
	}

	// Format and display results
	table := FormatTable(locations)
	_, _ = fmt.Fprint(deps.Stdout, table)

	return nil
}
