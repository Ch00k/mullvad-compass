package relays

import (
	"encoding/json"
	"testing"
)

func TestDetermineRelayType(t *testing.T) {
	tests := []struct {
		name         string
		endpointData string
		expected     ServerType
		shouldError  bool
	}{
		{
			name:         "String openvpn",
			endpointData: `"openvpn"`,
			expected:     OpenVPN,
			shouldError:  false,
		},
		{
			name:         "String bridge",
			endpointData: `"bridge"`,
			expected:     Bridge,
			shouldError:  false,
		},
		{
			name:         "Wireguard object",
			endpointData: `{"wireguard": {"public_key": "test123"}}`,
			expected:     WireGuard,
			shouldError:  false,
		},
		{
			name:         "Invalid format",
			endpointData: `{"unknown": "format"}`,
			expected:     ServerTypeNone,
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
					t.Errorf("Got type %v, expected %v", result, tt.expected)
				}
			}
		})
	}
}

func TestParseRelaysFile(t *testing.T) {
	// Test with the actual relays.json file
	relays, err := ParseRelaysFile("../../testdata/relays.json")
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
	relays, err := ParseRelaysFile("../../testdata/relays.json")
	if err != nil {
		t.Fatalf("Failed to parse relays.json: %v", err)
	}

	t.Run("All locations", func(t *testing.T) {
		locations, _, err := GetLocations(relays, ServerTypeNone, WGObfNone, false, IPv4)
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
		locations, _, err := GetLocations(relays, WireGuard, WGObfNone, false, IPv4)
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
		locations, _, err := GetLocations(relays, OpenVPN, WGObfNone, false, IPv4)
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
		locations, _, err := GetLocations(relays, ServerTypeNone, WGObfNone, false, IPv4)
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

	t.Run("Exclude relays with include_in_country=false", func(t *testing.T) {
		// Create test data with include_in_country field
		testRelays := &File{
			Countries: []Country{
				{
					Name: "Test Country",
					Code: "tc",
					Cities: []City{
						{
							Name:      "Test City",
							Code:      "test",
							Latitude:  50.0,
							Longitude: 10.0,
							Relays: []Relay{
								{
									Hostname:         "included-server",
									IPv4AddrIn:       "1.1.1.1",
									Active:           true,
									Owned:            true,
									Provider:         "test",
									IncludeInCountry: true,
									EndpointData:     json.RawMessage(`"openvpn"`),
									Location: RelayLocation{
										Country:   "Test Country",
										City:      "Test City",
										Latitude:  50.0,
										Longitude: 10.0,
									},
								},
								{
									Hostname:         "excluded-server",
									IPv4AddrIn:       "2.2.2.2",
									Active:           true,
									Owned:            true,
									Provider:         "test",
									IncludeInCountry: false,
									EndpointData:     json.RawMessage(`"openvpn"`),
									Location: RelayLocation{
										Country:   "Test Country",
										City:      "Test City",
										Latitude:  50.0,
										Longitude: 10.0,
									},
								},
							},
						},
					},
				},
			},
		}

		locations, _, err := GetLocations(testRelays, ServerTypeNone, WGObfNone, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		// Should only have the included server
		if len(locations) != 1 {
			t.Errorf("Expected 1 location, got %d", len(locations))
		}

		if len(locations) > 0 && locations[0].Hostname != "included-server" {
			t.Errorf("Expected included-server, got %s", locations[0].Hostname)
		}

		// Verify excluded server is not in results
		for _, loc := range locations {
			if loc.Hostname == "excluded-server" {
				t.Error("Found excluded-server which should have been filtered out")
			}
		}
	})

	t.Run("Filter by DAITA", func(t *testing.T) {
		locations, _, err := GetLocations(relays, ServerTypeNone, WGObfNone, true, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No DAITA servers found in test data")
		}

		// All returned servers should be wireguard with DAITA
		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server with DAITA filter: %s", loc.Hostname)
			}
		}
	})

	t.Run("Filter by LWO obfuscation", func(t *testing.T) {
		locations, _, err := GetLocations(relays, ServerTypeNone, LWO, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No LWO servers found in test data")
		}

		// All returned servers should be wireguard with LWO
		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server with LWO filter: %s", loc.Hostname)
			}
		}
	})

	t.Run("Filter by QUIC obfuscation", func(t *testing.T) {
		locations, _, err := GetLocations(relays, ServerTypeNone, QUIC, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No QUIC servers found in test data")
		}

		// All returned servers should be wireguard with QUIC
		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server with QUIC filter: %s", loc.Hostname)
			}
		}
	})

	t.Run("Filter by Shadowsocks obfuscation", func(t *testing.T) {
		locations, _, err := GetLocations(relays, ServerTypeNone, Shadowsocks, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No Shadowsocks servers found in test data")
		}

		// All returned servers should be wireguard with Shadowsocks
		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server with Shadowsocks filter: %s", loc.Hostname)
			}
		}
	})

	t.Run("Obfuscation filter with wireguard type", func(t *testing.T) {
		locations, _, err := GetLocations(relays, WireGuard, LWO, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		// All should be wireguard
		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server: %s", loc.Hostname)
			}
		}
	})

	t.Run("Obfuscation filter with openvpn type returns empty", func(t *testing.T) {
		locationsWithObf, _, err := GetLocations(relays, OpenVPN, LWO, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		// Should return 0 results (obfuscation forces wireguard, but serverType wants openvpn)
		if len(locationsWithObf) != 0 {
			t.Errorf("Expected 0 locations when filtering openvpn with wireguard obfuscation, got %d",
				len(locationsWithObf))
		}
	})

	t.Run("Filter locations without IPv6 when useIPv6 is true", func(t *testing.T) {
		testRelays := &File{
			Countries: []Country{
				{
					Name: "Test",
					Code: "TS",
					Cities: []City{
						{
							Name:      "Test City",
							Code:      "tst",
							Latitude:  50.0,
							Longitude: 10.0,
							Relays: []Relay{
								{
									Hostname:         "has-ipv6",
									IPv4AddrIn:       "1.1.1.1",
									IPv6AddrIn:       "2001:db8::1",
									Active:           true,
									Owned:            true,
									Provider:         "test",
									IncludeInCountry: true,
									EndpointData:     json.RawMessage(`"openvpn"`),
									Location: RelayLocation{
										Country:   "Test",
										City:      "Test City",
										Latitude:  50.0,
										Longitude: 10.0,
									},
								},
								{
									Hostname:         "no-ipv6",
									IPv4AddrIn:       "2.2.2.2",
									IPv6AddrIn:       "",
									Active:           true,
									Owned:            true,
									Provider:         "test",
									IncludeInCountry: true,
									EndpointData:     json.RawMessage(`"openvpn"`),
									Location: RelayLocation{
										Country:   "Test",
										City:      "Test City",
										Latitude:  50.0,
										Longitude: 10.0,
									},
								},
							},
						},
					},
				},
			},
		}

		locations, _, err := GetLocations(testRelays, ServerTypeNone, WGObfNone, false, IPv6)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		// Should only have the server with IPv6
		if len(locations) != 1 {
			t.Errorf("Expected 1 location with IPv6, got %d", len(locations))
		}

		if len(locations) > 0 && locations[0].Hostname != "has-ipv6" {
			t.Errorf("Expected has-ipv6, got %s", locations[0].Hostname)
		}

		// Verify no-ipv6 server is not in results
		for _, loc := range locations {
			if loc.Hostname == "no-ipv6" {
				t.Error("Found no-ipv6 which should have been filtered out")
			}
		}
	})

	t.Run("Include locations without IPv6 when useIPv6 is false", func(t *testing.T) {
		testRelays := &File{
			Countries: []Country{
				{
					Name: "Test",
					Code: "TS",
					Cities: []City{
						{
							Name:      "Test City",
							Code:      "tst",
							Latitude:  50.0,
							Longitude: 10.0,
							Relays: []Relay{
								{
									Hostname:         "has-ipv6",
									IPv4AddrIn:       "1.1.1.1",
									IPv6AddrIn:       "2001:db8::1",
									Active:           true,
									Owned:            true,
									Provider:         "test",
									IncludeInCountry: true,
									EndpointData:     json.RawMessage(`"openvpn"`),
									Location: RelayLocation{
										Country:   "Test",
										City:      "Test City",
										Latitude:  50.0,
										Longitude: 10.0,
									},
								},
								{
									Hostname:         "no-ipv6",
									IPv4AddrIn:       "2.2.2.2",
									IPv6AddrIn:       "",
									Active:           true,
									Owned:            true,
									Provider:         "test",
									IncludeInCountry: true,
									EndpointData:     json.RawMessage(`"openvpn"`),
									Location: RelayLocation{
										Country:   "Test",
										City:      "Test City",
										Latitude:  50.0,
										Longitude: 10.0,
									},
								},
							},
						},
					},
				},
			},
		}

		locations, _, err := GetLocations(testRelays, ServerTypeNone, WGObfNone, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		// Should have both servers
		if len(locations) != 2 {
			t.Errorf("Expected 2 locations, got %d", len(locations))
		}
	})
}
