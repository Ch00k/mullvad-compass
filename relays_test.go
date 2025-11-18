package main

import (
	"encoding/json"
	"testing"
)

func TestDetermineRelayType(t *testing.T) {
	tests := []struct {
		name         string
		endpointData string
		expected     string
		shouldError  bool
	}{
		{
			name:         "String openvpn",
			endpointData: `"openvpn"`,
			expected:     "openvpn",
			shouldError:  false,
		},
		{
			name:         "String bridge",
			endpointData: `"bridge"`,
			expected:     "bridge",
			shouldError:  false,
		},
		{
			name:         "Wireguard object",
			endpointData: `{"wireguard": {"public_key": "test123"}}`,
			expected:     "wireguard",
			shouldError:  false,
		},
		{
			name:         "Invalid format",
			endpointData: `{"unknown": "format"}`,
			expected:     "",
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw json.RawMessage
			if err := json.Unmarshal([]byte(tt.endpointData), &raw); err != nil {
				t.Fatalf("Failed to unmarshal test data: %v", err)
			}

			result, err := determineRelayType(raw)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Got type %q, expected %q", result, tt.expected)
				}
			}
		})
	}
}

func TestParseRelaysFile(t *testing.T) {
	// Test with the actual relays.json file
	relays, err := ParseRelaysFile("testdata/relays.json")
	if err != nil {
		t.Fatalf("Failed to parse relays.json: %v", err)
	}

	if len(relays.Countries) == 0 {
		t.Error("No countries found in relays.json")
	}

	hasRelays := false
	for _, country := range relays.Countries {
		for _, city := range country.Cities {
			if len(city.Relays) > 0 {
				hasRelays = true
				break
			}
		}
		if hasRelays {
			break
		}
	}

	if !hasRelays {
		t.Error("No relays found in any city")
	}
}

func TestGetLocations(t *testing.T) {
	relays, err := ParseRelaysFile("testdata/relays.json")
	if err != nil {
		t.Fatalf("Failed to parse relays.json: %v", err)
	}

	t.Run("All locations", func(t *testing.T) {
		locations, err := GetLocations(relays, "")
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Error("Expected some locations, got none")
		}

		// Verify no bridge servers are included
		for _, loc := range locations {
			if loc.Type == "bridge" {
				t.Errorf("Found bridge server which should be filtered out: %s", loc.Hostname)
			}
		}
	})

	t.Run("Filter by wireguard", func(t *testing.T) {
		locations, err := GetLocations(relays, "wireguard")
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server: %s (type: %s)", loc.Hostname, loc.Type)
			}
		}
	})

	t.Run("Filter by openvpn", func(t *testing.T) {
		locations, err := GetLocations(relays, "openvpn")
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		for _, loc := range locations {
			if loc.Type != "openvpn" {
				t.Errorf("Found non-openvpn server: %s (type: %s)", loc.Hostname, loc.Type)
			}
		}
	})

	t.Run("Verify location fields", func(t *testing.T) {
		locations, err := GetLocations(relays, "")
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No locations to verify")
		}

		loc := locations[0]
		if loc.IPv4Address == "" {
			t.Error("IPv4Address is empty")
		}
		if loc.Country == "" {
			t.Error("Country is empty")
		}
		if loc.City == "" {
			t.Error("City is empty")
		}
		if loc.Hostname == "" {
			t.Error("Hostname is empty")
		}
		if loc.Type == "" {
			t.Error("Type is empty")
		}
	})
}
