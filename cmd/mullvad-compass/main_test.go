package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Ch00k/mullvad-compass/internal/api"
	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

func TestE2E_FullFlow(t *testing.T) {
	t.Run("Successful execution with results", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  41.327953, // Tirana, Albania
					Longitude: 19.819025,
					Country:   "Albania",
					City:      "Tirana",
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				// Mock ping results - set latency for all locations
				for i := range locs {
					latency := float64(10.0 + float64(i)*5.0) // 10ms, 15ms, 20ms, etc.
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "100"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify output contains expected elements
		result := output.String()
		if !strings.Contains(result, "Country") {
			t.Error("Output should contain 'Country' header")
		}
		if !strings.Contains(result, "Latency") {
			t.Error("Output should contain 'Latency' header")
		}
		if !strings.Contains(result, "Albania") {
			t.Error("Output should contain Albania servers (within 100km of Tirana)")
		}

		// Should have table structure
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) < 3 {
			t.Errorf("Expected at least 3 lines (header, separator, data), got %d", len(lines))
		}
	})

	t.Run("Filter by server type", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  52.520008, // Berlin
					Longitude: 13.404954,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				for i := range locs {
					latency := 15.0
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-s", "wireguard", "-m", "500"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()

		// Should only contain wireguard servers
		if strings.Contains(result, "openvpn") {
			t.Error("Output should not contain openvpn servers when filtering for wireguard")
		}
	})

	t.Run("No servers within distance", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  -90.0, // South Pole
					Longitude: 0.0,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "100"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "No servers found within") {
			t.Error("Should indicate no servers found within distance")
		}
	})

	t.Run("Locations sorted by latency", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  50.0,
					Longitude: 10.0,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				// Set different latencies in reverse order
				for i := range locs {
					latency := float64(100 - i*10) // 100ms, 90ms, 80ms, etc.
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "2000"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Output should contain formatted table
		result := output.String()
		if !strings.Contains(result, "---") {
			t.Error("Output should contain table separator")
		}
	})

	t.Run("Handle ping timeouts", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  50.0,
					Longitude: 10.0,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				// Set some locations with nil latency (timeout)
				for i := range locs {
					if i%2 == 0 {
						locs[i].Latency = nil // timeout
					} else {
						latency := 25.0
						locs[i].Latency = &latency
					}
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "1000"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "timeout") {
			t.Error("Output should contain 'timeout' for nil latency values")
		}
	})
}

func TestE2E_ErrorHandling(t *testing.T) {
	t.Run("API error", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return nil, fmt.Errorf("API connection failed")
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "500"}
		err := run(context.Background(), args, deps)

		if err == nil {
			t.Fatal("Expected error when API fails")
		}
		if !strings.Contains(err.Error(), "failed to get user location") {
			t.Errorf("Expected user location error, got: %v", err)
		}
	})

	t.Run("Relays file not found", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{Latitude: 50.0, Longitude: 10.0}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return nil, fmt.Errorf("file not found: nonexistent.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "500"}
		err := run(context.Background(), args, deps)

		if err == nil {
			t.Fatal("Expected error when relays file not found")
		}
	})

	t.Run("Relays path detection error", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{Latitude: 50.0, Longitude: 10.0}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return nil, fmt.Errorf("unsupported platform")
			},
			Stdout: &output,
		}

		args := []string{"-m", "500"}
		err := run(context.Background(), args, deps)

		if err == nil {
			t.Fatal("Expected error when relays path cannot be detected")
		}
		if !strings.Contains(err.Error(), "unsupported platform") {
			t.Errorf("Expected unsupported platform error, got: %v", err)
		}
	})

	t.Run("Connected to Mullvad VPN warning", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:      59.329323,
					Longitude:     18.068581,
					Country:       "Sweden",
					City:          "Stockholm",
					MullvadExitIP: true,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				t.Error("PingLocations should not be called when connected to Mullvad VPN")
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "500"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "connected to Mullvad VPN") {
			t.Error("Output should warn about being connected to Mullvad VPN")
		}
	})
}

func TestE2E_FlagParsing(t *testing.T) {
	t.Run("Invalid server type", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				t.Error("Should not call GetUserLocation with invalid flags")
				return nil, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-s", "invalid"}
		err := run(context.Background(), args, deps)

		if err == nil {
			t.Fatal("Expected error for invalid server type")
		}
		if !strings.Contains(err.Error(), "invalid server type") {
			t.Errorf("Expected invalid server type error, got: %v", err)
		}
	})

	t.Run("Invalid max distance", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				t.Error("Should not call GetUserLocation with invalid flags")
				return nil, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "invalid"}
		err := run(context.Background(), args, deps)

		if err == nil {
			t.Fatal("Expected error for invalid max distance")
		}
	})

	t.Run("Negative max distance", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				t.Error("Should not call GetUserLocation with invalid flags")
				return nil, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-m", "-100"}
		err := run(context.Background(), args, deps)

		if err == nil {
			t.Fatal("Expected error for negative max distance")
		}
		if !strings.Contains(err.Error(), "must be positive") {
			t.Errorf("Expected positive distance error, got: %v", err)
		}
	})

	t.Run("Unknown flag", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				t.Error("Should not call GetUserLocation with invalid flags")
				return nil, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"--unknown-flag"}
		err := run(context.Background(), args, deps)

		if err == nil {
			t.Fatal("Expected error for unknown flag")
		}
		if !strings.Contains(err.Error(), "unknown flag") {
			t.Errorf("Expected unknown flag error, got: %v", err)
		}
	})
}

func TestE2E_BestServerMode(t *testing.T) {
	t.Run("No arguments triggers best server mode with progressive range expansion", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  -90.0, // South Pole - no servers nearby
					Longitude: 0.0,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				// Set different latencies
				for i := range locs {
					latency := float64(20 + i*5) // 20ms, 25ms, 30ms, etc.
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{} // No arguments - should trigger best server mode
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()

		// Should use compact format similar to --where-am-i
		if !strings.Contains(result, "Hostname:") {
			t.Error("Output should contain 'Hostname:' field")
		}
		if !strings.Contains(result, "Country:") {
			t.Error("Output should contain 'Country:' field")
		}
		if !strings.Contains(result, "City:") {
			t.Error("Output should contain 'City:' field")
		}
		if !strings.Contains(result, "Latency:") {
			t.Error("Output should contain 'Latency:' field")
		}
		if !strings.Contains(result, "Distance:") {
			t.Error("Output should contain 'Distance:' field")
		}

		// Should contain the best server (lowest latency)
		if !strings.Contains(result, "20.00") {
			t.Error("Should contain the server with lowest latency (20ms)")
		}
	})

	t.Run("Best server mode finds servers in progressively expanding ranges", func(t *testing.T) {
		var output bytes.Buffer
		callCount := 0

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  0.0,
					Longitude: 0.0,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				callCount++
				for i := range locs {
					latency := float64(30 + i*2)
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should have called PingLocations at least once
		if callCount == 0 {
			t.Error("Expected PingLocations to be called at least once")
		}

		// Should output one server in compact format
		result := output.String()
		if !strings.Contains(result, "Hostname:") {
			t.Error("Output should contain 'Hostname:' field")
		}
		if !strings.Contains(result, "Latency:") {
			t.Error("Output should contain 'Latency:' field")
		}
	})

	t.Run("Best server mode stops at maximum radius to prevent infinite loop", func(t *testing.T) {
		var output bytes.Buffer
		tempFile, err := os.CreateTemp("", "far-relays-*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Remove(tempFile.Name()); err != nil {
				t.Logf("failed to remove temp file: %v", err)
			}
		}()

		farRelays := `{
			"etag": "test",
			"countries": [
				{
					"name": "Antarctica",
					"code": "aq",
					"cities": [
						{
							"name": "South Pole",
							"code": "spo",
							"latitude": -90.0,
							"longitude": 0.0,
							"relays": [
								{
									"hostname": "aq-spo-wg-001",
									"ipv4_addr_in": "192.0.2.1",
									"ipv6_addr_in": "2001:db8::1",
									"include_in_country": true,
									"active": true,
									"owned": true,
									"provider": "test",
									"weight": 1,
									"endpoint_data": {
										"wireguard": {
											"public_key": "test",
											"daita": false
										}
									},
									"location": {
										"country": "Antarctica",
										"country_code": "aq",
										"city": "South Pole",
										"city_code": "spo",
										"latitude": -90.0,
										"longitude": 0.0
									}
								}
							]
						}
					]
				}
			]
		}`
		if _, err := tempFile.Write([]byte(farRelays)); err != nil {
			t.Fatal(err)
		}
		if err := tempFile.Close(); err != nil {
			t.Fatal(err)
		}

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  90.0,
					Longitude: 0.0,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				for i := range locs {
					latency := float64(30 + i*2)
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile(tempFile.Name())
			},
			Stdout: &output,
		}

		args := []string{}
		err = run(context.Background(), args, deps)
		if err == nil {
			t.Fatal("Expected error when no servers found within maximum radius, got nil")
		}

		expectedError := "no servers found within maximum search radius"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error containing %q, got: %v", expectedError, err)
		}
	})

	t.Run("Any argument disables best server mode", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  50.0,
					Longitude: 10.0,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				for i := range locs {
					latency := float64(15 + i*3)
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		// With any argument, should use normal mode
		args := []string{"-m", "500"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()
		lines := strings.Split(strings.TrimSpace(result), "\n")
		// Should have more than 3 lines (multiple servers)
		if len(lines) <= 3 {
			t.Error("With arguments, should show multiple servers, not just best one")
		}
	})
}

func TestE2E_Integration(t *testing.T) {
	t.Run("Multiple filters combined", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  50.110924, // Frankfurt
					Longitude: 8.682127,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				for i := range locs {
					latency := 12.5
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return relays.ParseRelaysFile("../../testdata/relays.json")
			},
			Stdout: &output,
		}

		args := []string{"-s", "wireguard", "-m", "300"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()

		// Should have table output
		if !strings.Contains(result, "Country") {
			t.Error("Should contain table header")
		}

		// All results should be wireguard
		lines := strings.Split(result, "\n")
		for i, line := range lines {
			if i < 2 { // Skip header and separator
				continue
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			if strings.Contains(line, "openvpn") {
				t.Error("Should not contain openvpn servers")
			}
		}
	})

	t.Run("WhereAmI flag shows location with long form not on Mullvad", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					IP:            "203.0.113.42",
					Latitude:      41.327953,
					Longitude:     19.819025,
					Country:       "Albania",
					City:          "Tirana",
					MullvadExitIP: false,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				t.Error("PingLocations should not be called with --where-am-i flag")
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				t.Error("GetRelaysPath should not be called with --where-am-i flag")
				return nil, fmt.Errorf("test error")
			},
			Stdout: &output,
		}

		args := []string{"--where-am-i"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()

		// Should contain location information
		if !strings.Contains(result, "203.0.113.42") {
			t.Error("Output should contain IP address")
		}
		if !strings.Contains(result, "Tirana") {
			t.Error("Output should contain city name")
		}
		if !strings.Contains(result, "Albania") {
			t.Error("Output should contain country name")
		}
		if !strings.Contains(result, "41.327953") {
			t.Error("Output should contain latitude")
		}
		if !strings.Contains(result, "19.819025") {
			t.Error("Output should contain longitude")
		}
		if !strings.Contains(result, "Connected to Mullvad VPN") || !strings.Contains(result, "No") {
			t.Error("Output should indicate not connected to Mullvad VPN")
		}
	})

	t.Run("WhereAmI flag shows Mullvad connection status when connected", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					IP:            "185.65.135.42",
					Latitude:      59.329323,
					Longitude:     18.068581,
					Country:       "Sweden",
					City:          "Stockholm",
					MullvadExitIP: true,
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				t.Error("PingLocations should not be called with --where-am-i flag")
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				t.Error("GetRelaysPath should not be called with --where-am-i flag")
				return nil, fmt.Errorf("test error")
			},
			Stdout: &output,
		}

		args := []string{"--where-am-i"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()

		if !strings.Contains(result, "185.65.135.42") {
			t.Error("Output should contain IP address")
		}
		if !strings.Contains(result, "Connected to Mullvad VPN") || !strings.Contains(result, "Yes") {
			t.Error("Output should indicate connected to Mullvad VPN")
		}
	})

	t.Run("WhereAmI flag shows location with short form", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  52.520008,
					Longitude: 13.404954,
					Country:   "Germany",
					City:      "Berlin",
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				t.Error("PingLocations should not be called with -i flag")
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				t.Error("GetRelaysPath should not be called with -i flag")
				return nil, fmt.Errorf("test error")
			},
			Stdout: &output,
		}

		args := []string{"-i"}
		err := run(context.Background(), args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()

		// Should contain location information
		if !strings.Contains(result, "Berlin") {
			t.Error("Output should contain city name")
		}
		if !strings.Contains(result, "Germany") {
			t.Error("Output should contain country name")
		}
		if !strings.Contains(result, "52.520008") {
			t.Error("Output should contain latitude")
		}
		if !strings.Contains(result, "13.404954") {
			t.Error("Output should contain longitude")
		}
	})
}
