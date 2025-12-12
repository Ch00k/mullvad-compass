package main

import (
	"testing"
)

func TestGetDeterministicUserLocation(t *testing.T) {
	loc := getDeterministicUserLocation()

	t.Run("Has expected IP", func(t *testing.T) {
		if loc.IP != "203.0.113.42" {
			t.Errorf("Expected IP 203.0.113.42, got %s", loc.IP)
		}
	})

	t.Run("Has expected country", func(t *testing.T) {
		if loc.Country != "Germany" {
			t.Errorf("Expected country Germany, got %s", loc.Country)
		}
	})

	t.Run("Has expected city", func(t *testing.T) {
		if loc.City != "Dresden" {
			t.Errorf("Expected city Dresden, got %s", loc.City)
		}
	})

	t.Run("Has expected coordinates", func(t *testing.T) {
		expectedLat := 51.0514
		expectedLon := 13.7341
		if loc.Latitude != expectedLat {
			t.Errorf("Expected latitude %f, got %f", expectedLat, loc.Latitude)
		}
		if loc.Longitude != expectedLon {
			t.Errorf("Expected longitude %f, got %f", expectedLon, loc.Longitude)
		}
	})

	t.Run("Is not Mullvad exit IP", func(t *testing.T) {
		if loc.MullvadExitIP {
			t.Error("Expected MullvadExitIP to be false")
		}
	})
}

func TestGetDeterministicLocations(t *testing.T) {
	locs := getDeterministicLocations()

	t.Run("Returns expected number of locations", func(t *testing.T) {
		expected := 11
		if len(locs) != expected {
			t.Errorf("Expected %d locations, got %d", expected, len(locs))
		}
	})

	t.Run("All locations have required fields", func(t *testing.T) {
		for i, loc := range locs {
			if loc.Country == "" {
				t.Errorf("Location %d missing Country", i)
			}
			if loc.City == "" {
				t.Errorf("Location %d missing City", i)
			}
			if loc.Type == "" {
				t.Errorf("Location %d missing Type", i)
			}
			if loc.Hostname == "" {
				t.Errorf("Location %d missing Hostname", i)
			}
			if loc.IPv4Address == "" {
				t.Errorf("Location %d missing IPv4Address", i)
			}
			if loc.DistanceFromMyLocation == nil {
				t.Errorf("Location %d missing DistanceFromMyLocation", i)
			}
			if loc.Latency == nil {
				t.Errorf("Location %d missing Latency", i)
			}
		}
	})

	t.Run("All locations are wireguard type", func(t *testing.T) {
		for i, loc := range locs {
			if loc.Type != "wireguard" {
				t.Errorf("Location %d has type %s, expected wireguard", i, loc.Type)
			}
		}
	})

	t.Run("Locations are sorted by latency", func(t *testing.T) {
		for i := 1; i < len(locs); i++ {
			if *locs[i-1].Latency > *locs[i].Latency {
				t.Errorf("Locations not sorted by latency: location %d has %f ms, location %d has %f ms",
					i-1, *locs[i-1].Latency, i, *locs[i].Latency)
			}
		}
	})

	t.Run("Has expected countries", func(t *testing.T) {
		czechCount := 0
		germanyCount := 0
		for _, loc := range locs {
			switch loc.Country {
			case "Czech Republic":
				czechCount++
			case "Germany":
				germanyCount++
			default:
				t.Errorf("Unexpected country: %s", loc.Country)
			}
		}
		if czechCount != 3 {
			t.Errorf("Expected 3 Czech Republic locations, got %d", czechCount)
		}
		if germanyCount != 8 {
			t.Errorf("Expected 8 Germany locations, got %d", germanyCount)
		}
	})

	t.Run("Has expected cities", func(t *testing.T) {
		pragueCount := 0
		berlinCount := 0
		for _, loc := range locs {
			switch loc.City {
			case "Prague":
				pragueCount++
			case "Berlin":
				berlinCount++
			default:
				t.Errorf("Unexpected city: %s", loc.City)
			}
		}
		if pragueCount != 3 {
			t.Errorf("Expected 3 Prague locations, got %d", pragueCount)
		}
		if berlinCount != 8 {
			t.Errorf("Expected 8 Berlin locations, got %d", berlinCount)
		}
	})

	t.Run("Has expected distances", func(t *testing.T) {
		distance156Count := 0
		distance238Count := 0
		for _, loc := range locs {
			switch *loc.DistanceFromMyLocation {
			case 156.0:
				distance156Count++
			case 238.0:
				distance238Count++
			default:
				t.Errorf("Unexpected distance: %f", *loc.DistanceFromMyLocation)
			}
		}
		if distance156Count != 3 {
			t.Errorf("Expected 3 locations at 156km, got %d", distance156Count)
		}
		if distance238Count != 8 {
			t.Errorf("Expected 8 locations at 238km, got %d", distance238Count)
		}
	})

	t.Run("Best server is Prague with lowest latency", func(t *testing.T) {
		best := locs[0]
		if best.City != "Prague" {
			t.Errorf("Expected best server to be in Prague, got %s", best.City)
		}
		if best.Hostname != "cz-prg-wg-201" {
			t.Errorf("Expected best server hostname cz-prg-wg-201, got %s", best.Hostname)
		}
		expected := 9.78
		if *best.Latency != expected {
			t.Errorf("Expected best server latency %f, got %f", expected, *best.Latency)
		}
	})

	t.Run("All hostnames are unique", func(t *testing.T) {
		hostnames := make(map[string]bool)
		for _, loc := range locs {
			if hostnames[loc.Hostname] {
				t.Errorf("Duplicate hostname: %s", loc.Hostname)
			}
			hostnames[loc.Hostname] = true
		}
	})

	t.Run("All IPv4 addresses are unique", func(t *testing.T) {
		ips := make(map[string]bool)
		for _, loc := range locs {
			if ips[loc.IPv4Address] {
				t.Errorf("Duplicate IPv4 address: %s", loc.IPv4Address)
			}
			ips[loc.IPv4Address] = true
		}
	})
}
