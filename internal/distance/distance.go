// Package distance provides functions for calculating and filtering locations by distance.
package distance

import (
	"log"
	"math"

	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

const earthRadiusKm = 6371.0

// CalculateDistance computes the geodesic distance between two points using the Haversine formula
// Returns distance in kilometers
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)
	deltaLat := degreesToRadians(lat2 - lat1)
	deltaLon := degreesToRadians(lon2 - lon1)

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// degreesToRadians converts degrees to radians
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

// FilterByDistance returns locations within the specified distance threshold
func FilterByDistance(locations []relays.Location, userLat, userLon, maxDistance float64) []relays.Location {
	return FilterByDistanceWithLogLevel(locations, userLat, userLon, maxDistance, logging.LogLevelError)
}

// FilterByDistanceWithLogLevel returns locations within the specified distance threshold with logging support
func FilterByDistanceWithLogLevel(
	locations []relays.Location,
	userLat, userLon, maxDistance float64,
	logLevel logging.LogLevel,
) []relays.Location {
	if logLevel <= logging.LogLevelDebug {
		log.Printf(
			"Filtering %d locations within %.1f km of user location (%.4f, %.4f)",
			len(locations),
			maxDistance,
			userLat,
			userLon,
		)
	}

	var filtered []relays.Location

	for _, loc := range locations {
		distance := CalculateDistance(userLat, userLon, loc.Latitude, loc.Longitude)
		if distance <= maxDistance {
			d := distance // Allocate new variable to avoid pointer aliasing
			loc.DistanceFromMyLocation = &d
			filtered = append(filtered, loc)
		}
	}

	if logLevel <= logging.LogLevelInfo {
		log.Printf(
			"Filtered to %d locations within %.1f km (filtered out %d)",
			len(filtered),
			maxDistance,
			len(locations)-len(filtered),
		)
	}

	return filtered
}
