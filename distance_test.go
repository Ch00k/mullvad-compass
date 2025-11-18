package main

import (
	"math"
	"testing"
)

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name      string
		lat1      float64
		lon1      float64
		lat2      float64
		lon2      float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "Same location",
			lat1:      50.0,
			lon1:      10.0,
			lat2:      50.0,
			lon2:      10.0,
			expected:  0.0,
			tolerance: 0.1,
		},
		{
			name:      "New York to London",
			lat1:      40.7128,
			lon1:      -74.0060,
			lat2:      51.5074,
			lon2:      -0.1278,
			expected:  5570.0,
			tolerance: 10.0,
		},
		{
			name:      "Sydney to Tokyo",
			lat1:      -33.8688,
			lon1:      151.2093,
			lat2:      35.6762,
			lon2:      139.6503,
			expected:  7823.0,
			tolerance: 10.0,
		},
		{
			name:      "Short distance",
			lat1:      48.8566,
			lon1:      2.3522,
			lat2:      48.8606,
			lon2:      2.3376,
			expected:  1.1,
			tolerance: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			diff := math.Abs(result - tt.expected)
			if diff > tt.tolerance {
				t.Errorf("CalculateDistance() = %.2f, expected %.2f (±%.2f), diff = %.2f",
					result, tt.expected, tt.tolerance, diff)
			}
		})
	}
}

func TestFilterByDistance(t *testing.T) {
	userLat := 50.0
	userLon := 10.0

	locations := []Location{
		{Hostname: "close", Latitude: 50.1, Longitude: 10.1},
		{Hostname: "far", Latitude: 60.0, Longitude: 20.0},
		{Hostname: "medium", Latitude: 51.0, Longitude: 11.0},
	}

	filtered := FilterByDistance(locations, userLat, userLon, 200.0)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 locations within 200km, got %d", len(filtered))
	}

	for _, loc := range filtered {
		if loc.DistanceFromMyLocation == nil {
			t.Errorf("Location %s missing distance value", loc.Hostname)
		}
		if *loc.DistanceFromMyLocation > 200.0 {
			t.Errorf("Location %s has distance %.2f km, exceeds threshold of 200km",
				loc.Hostname, *loc.DistanceFromMyLocation)
		}
	}
}

func TestDegreesToRadians(t *testing.T) {
	tests := []struct {
		degrees  float64
		expected float64
	}{
		{0, 0},
		{90, math.Pi / 2},
		{180, math.Pi},
		{360, 2 * math.Pi},
	}

	for _, tt := range tests {
		result := degreesToRadians(tt.degrees)
		if math.Abs(result-tt.expected) > 0.0001 {
			t.Errorf("degreesToRadians(%.0f) = %f, expected %f", tt.degrees, result, tt.expected)
		}
	}
}
