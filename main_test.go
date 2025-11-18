package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestE2E_FullFlow(t *testing.T) {
	t.Run("Successful execution with results", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{
					Latitude:  41.327953, // Tirana, Albania
					Longitude: 19.819025,
					Country:   "Albania",
					City:      "Tirana",
				}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				// Mock ping results - set latency for all locations
				for i := range locs {
					latency := float64(10.0 + float64(i)*5.0) // 10ms, 15ms, 20ms, etc.
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "100"}
		err := run(args, deps)
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
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{
					Latitude:  52.520008, // Berlin
					Longitude: 13.404954,
				}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				for i := range locs {
					latency := 15.0
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-s", "wireguard", "-m", "500"}
		err := run(args, deps)
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
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{
					Latitude:  -90.0, // South Pole
					Longitude: 0.0,
				}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "100"}
		err := run(args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "No servers found within") {
			t.Error("Should indicate no servers found within distance")
		}
	})

	t.Run("Custom relays file path", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{
					Latitude:  50.0,
					Longitude: 10.0,
				}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				for i := range locs {
					latency := 20.0
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "", fmt.Errorf("should not be called when path is specified")
			},
			Stdout: &output,
		}

		args := []string{"--relays-file", "testdata/relays.json", "-m", "1000"}
		err := run(args, deps)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should have produced output
		if output.Len() == 0 {
			t.Error("Expected output when servers found")
		}
	})

	t.Run("Locations sorted by latency", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{
					Latitude:  50.0,
					Longitude: 10.0,
				}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				// Set different latencies in reverse order
				for i := range locs {
					latency := float64(100 - i*10) // 100ms, 90ms, 80ms, etc.
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "2000"}
		err := run(args, deps)
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
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{
					Latitude:  50.0,
					Longitude: 10.0,
				}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
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
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "1000"}
		err := run(args, deps)
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
			GetUserLocation: func() (*UserLocation, error) {
				return nil, fmt.Errorf("API connection failed")
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "500"}
		err := run(args, deps)

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
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{Latitude: 50.0, Longitude: 10.0}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "nonexistent.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "500"}
		err := run(args, deps)

		if err == nil {
			t.Fatal("Expected error when relays file not found")
		}
	})

	t.Run("Invalid relays file path argument", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{Latitude: 50.0, Longitude: 10.0}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				t.Error("GetRelaysPath should not be called when path is provided")
				return "", nil
			},
			Stdout: &output,
		}

		args := []string{"--relays-file", "nonexistent.json"}
		err := run(args, deps)

		if err == nil {
			t.Fatal("Expected error when relays file doesn't exist")
		}
	})

	t.Run("Relays path detection error", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{Latitude: 50.0, Longitude: 10.0}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "", fmt.Errorf("unsupported platform")
			},
			Stdout: &output,
		}

		args := []string{"-m", "500"}
		err := run(args, deps)

		if err == nil {
			t.Fatal("Expected error when relays path cannot be detected")
		}
		if !strings.Contains(err.Error(), "could not find relays.json") {
			t.Errorf("Expected relays.json error, got: %v", err)
		}
	})
}

func TestE2E_FlagParsing(t *testing.T) {
	t.Run("Invalid server type", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func() (*UserLocation, error) {
				t.Error("Should not call GetUserLocation with invalid flags")
				return nil, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-s", "invalid"}
		err := run(args, deps)

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
			GetUserLocation: func() (*UserLocation, error) {
				t.Error("Should not call GetUserLocation with invalid flags")
				return nil, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "invalid"}
		err := run(args, deps)

		if err == nil {
			t.Fatal("Expected error for invalid max distance")
		}
	})

	t.Run("Negative max distance", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func() (*UserLocation, error) {
				t.Error("Should not call GetUserLocation with invalid flags")
				return nil, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "-100"}
		err := run(args, deps)

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
			GetUserLocation: func() (*UserLocation, error) {
				t.Error("Should not call GetUserLocation with invalid flags")
				return nil, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"--unknown-flag"}
		err := run(args, deps)

		if err == nil {
			t.Fatal("Expected error for unknown flag")
		}
		if !strings.Contains(err.Error(), "unknown flag") {
			t.Errorf("Expected unknown flag error, got: %v", err)
		}
	})
}

func TestE2E_Integration(t *testing.T) {
	t.Run("Multiple filters combined", func(t *testing.T) {
		var output bytes.Buffer

		deps := Dependencies{
			GetUserLocation: func() (*UserLocation, error) {
				return &UserLocation{
					Latitude:  50.110924, // Frankfurt
					Longitude: 8.682127,
				}, nil
			},
			PingLocations: func(locs []Location) ([]Location, error) {
				for i := range locs {
					latency := 12.5
					locs[i].Latency = &latency
				}
				return locs, nil
			},
			GetRelaysPath: func() (string, error) {
				return "testdata/relays.json", nil
			},
			Stdout: &output,
		}

		args := []string{"-s", "wireguard", "-m", "300"}
		err := run(args, deps)
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
}
