package main

import (
	"context"
	"log"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/api"
	"github.com/Ch00k/mullvad-compass/internal/distance"
	"github.com/Ch00k/mullvad-compass/internal/formatter"
	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// getUserLocation fetches user location with optional debug timing
func getUserLocation(
	ctx context.Context,
	logLevel logging.LogLevel,
	getUserLocationFn func(context.Context, logging.LogLevel) (*api.UserLocation, error),
) (*api.UserLocation, error) {
	start := time.Now()
	defer func() {
		if logLevel <= logging.LogLevelDebug {
			elapsed := time.Since(start)
			log.Printf("User location fetch completed in %v", elapsed)
		}
	}()

	return getUserLocationFn(ctx, logLevel)
}

// parseRelaysFile parses the relays JSON file with optional debug timing
// If path is empty, it will attempt to find the default relays.json location
func parseRelaysFile(
	logLevel logging.LogLevel,
	path string,
	getRelaysPathFn func() (string, error),
) (*relays.File, error) {
	start := time.Now()
	defer func() {
		if logLevel <= logging.LogLevelDebug {
			elapsed := time.Since(start)
			log.Printf("Parse relays file completed in %v", elapsed)
		}
	}()

	// If no path provided, try to find default location
	if path == "" {
		defaultPath, err := getRelaysPathFn()
		if err != nil {
			return nil, err
		}
		path = defaultPath
	}

	return relays.ParseRelaysFileWithLogLevel(path, logLevel)
}

// getLocations fetches and filters relay locations with optional debug timing
func getLocations(
	logLevel logging.LogLevel,
	relaysData *relays.File,
	antiCensorship relays.AntiCensorship,
	daita bool,
	ipVersion relays.IPVersion,
) ([]relays.Location, error) {
	start := time.Now()
	defer func() {
		if logLevel <= logging.LogLevelDebug {
			elapsed := time.Since(start)
			log.Printf("Get locations completed in %v", elapsed)
		}
	}()

	locations, skipped, err := relays.GetLocations(relaysData, antiCensorship, daita, ipVersion)
	if err != nil {
		return nil, err
	}

	if skipped > 0 && logLevel <= logging.LogLevelWarning {
		log.Printf("Warning: %d relay(s) skipped due to unknown endpoint_data format", skipped)
	}

	return locations, nil
}

// filterByDistance filters locations by distance with optional debug timing
func filterByDistance(
	logLevel logging.LogLevel,
	locations []relays.Location,
	userLat, userLon, maxDistance float64,
) []relays.Location {
	start := time.Now()
	defer func() {
		if logLevel <= logging.LogLevelDebug {
			elapsed := time.Since(start)
			log.Printf("Filter by distance completed in %v", elapsed)
		}
	}()

	return distance.FilterByDistanceWithLogLevel(locations, userLat, userLon, maxDistance, logLevel)
}

// pingLocations pings locations with optional debug timing
func pingLocations(
	ctx context.Context,
	logLevel logging.LogLevel,
	locations []relays.Location,
	timeout, workers int,
	ipVersion relays.IPVersion,
	pingLocationsFn func(context.Context, []relays.Location, int, int, relays.IPVersion, logging.LogLevel) ([]relays.Location, error),
) ([]relays.Location, error) {
	start := time.Now()
	defer func() {
		if logLevel <= logging.LogLevelDebug {
			elapsed := time.Since(start)
			log.Printf("Ping locations completed in %v", elapsed)
		}
	}()

	return pingLocationsFn(ctx, locations, timeout, workers, ipVersion, logLevel)
}

// sortLocationsByLatency sorts locations by latency with optional debug timing
func sortLocationsByLatency(
	logLevel logging.LogLevel,
	locations []relays.Location,
) {
	start := time.Now()
	defer func() {
		if logLevel <= logging.LogLevelDebug {
			elapsed := time.Since(start)
			log.Printf("Sort locations by latency completed in %v", elapsed)
		}
	}()

	formatter.SortLocationsByLatency(locations)
}
