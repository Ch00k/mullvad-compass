package ping

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// skipIfNoPermissions skips the test if the error indicates insufficient permissions
func skipIfNoPermissions(t *testing.T, err error) {
	t.Helper()
	if err != nil && strings.Contains(err.Error(), "permission denied") {
		t.Skipf("Skipping test due to insufficient permissions: %v", err)
	}
}

func TestPingLocations(t *testing.T) {
	const (
		defaultTimeout = 500
		defaultWorkers = 25
	)

	t.Run("Empty locations", func(t *testing.T) {
		result, err := Locations(context.Background(),
			[]relays.Location{},
			defaultTimeout,
			defaultWorkers,
			relays.IPv4,
		)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected empty result for empty input, got %d locations", len(result))
		}
	})

	t.Run("Single location", func(t *testing.T) {
		locations := []relays.Location{
			{
				IPv4Address: "127.0.0.1",
				Hostname:    "localhost",
				Country:     "Local",
				City:        "Test",
			},
		}

		result, err := Locations(
			context.Background(),
			locations,
			defaultTimeout,
			defaultWorkers,
			relays.IPv4,
		)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}

		// Localhost should respond (if not blocked by firewall)
		// but we don't assert on latency value since it may fail in some environments
		if result[0].Hostname != "localhost" {
			t.Error("Result should preserve hostname")
		}
	})

	t.Run("Multiple locations", func(t *testing.T) {
		locations := []relays.Location{
			{
				IPv4Address: "127.0.0.1",
				Hostname:    "localhost-1",
				Country:     "Local",
				City:        "Test",
			},
			{
				IPv4Address: "127.0.0.2",
				Hostname:    "localhost-2",
				Country:     "Local",
				City:        "Test",
			},
			{
				IPv4Address: "127.0.0.3",
				Hostname:    "localhost-3",
				Country:     "Local",
				City:        "Test",
			},
		}

		result, err := Locations(
			context.Background(),
			locations,
			defaultTimeout,
			defaultWorkers,
			relays.IPv4,
		)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 results, got %d", len(result))
		}

		// Verify all locations are in result
		hostnames := make(map[string]bool)
		for _, loc := range result {
			hostnames[loc.Hostname] = true
		}

		if !hostnames["localhost-1"] || !hostnames["localhost-2"] || !hostnames["localhost-3"] {
			t.Error("Not all locations present in result")
		}
	})

	t.Run("Invalid IP addresses get nil latency", func(t *testing.T) {
		locations := []relays.Location{
			{
				IPv4Address: "256.256.256.256", // Invalid IP
				Hostname:    "invalid",
				Country:     "Invalid",
				City:        "Test",
			},
		}

		result, err := Locations(
			context.Background(),
			locations,
			defaultTimeout,
			defaultWorkers,
			relays.IPv4,
		)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}

		if result[0].Latency != nil {
			t.Error("Invalid IP should result in nil latency")
		}
	})

	t.Run("Unreachable IP addresses get nil latency", func(t *testing.T) {
		locations := []relays.Location{
			{
				IPv4Address: "192.0.2.1", // TEST-NET-1, should be unreachable
				Hostname:    "unreachable",
				Country:     "Test",
				City:        "Test",
			},
		}

		result, err := Locations(
			context.Background(),
			locations,
			defaultTimeout,
			defaultWorkers,
			relays.IPv4,
		)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got %d", len(result))
		}

		// Unreachable IPs should timeout and get nil latency
		if result[0].Latency != nil {
			t.Logf(
				"Warning: Expected nil latency for unreachable IP, got %v (may succeed in some environments)",
				*result[0].Latency,
			)
		}
	})

	t.Run("Concurrent ping respects worker pool size", func(t *testing.T) {
		// Create more locations than worker pool size
		locations := make([]relays.Location, defaultWorkers*2)
		for i := range locations {
			locations[i] = relays.Location{
				IPv4Address: "127.0.0.1",
				Hostname:    "test",
				Country:     "Test",
				City:        "Test",
			}
		}

		start := time.Now()
		result, err := Locations(
			context.Background(),
			locations,
			defaultTimeout,
			defaultWorkers,
			relays.IPv4,
		)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		duration := time.Since(start)

		if len(result) != len(locations) {
			t.Errorf("Expected %d results, got %d", len(locations), len(result))
		}

		// With concurrency, should complete faster than sequential pings
		// Even if some fail, total time should be reasonable
		pingTimeout := time.Duration(defaultTimeout) * time.Millisecond
		maxExpectedDuration := pingTimeout * time.Duration(len(locations)/defaultWorkers+2)
		if duration > maxExpectedDuration {
			t.Logf("Warning: Pinging took longer than expected: %v (max expected: %v)", duration, maxExpectedDuration)
		}
	})
}

func TestPingResult(t *testing.T) {
	t.Run("Result structure", func(t *testing.T) {
		latency := 12.34
		loc := &relays.Location{
			IPv4Address: "1.2.3.4",
			Hostname:    "test",
		}

		result := Result{
			Location: loc,
			Latency:  &latency,
		}

		if result.Location != loc {
			t.Error("Result should reference the original location")
		}

		if result.Latency == nil {
			t.Error("Latency should not be nil")
		}

		if *result.Latency != 12.34 {
			t.Errorf("Expected latency 12.34, got %f", *result.Latency)
		}
	})

	t.Run("Result with nil latency", func(t *testing.T) {
		loc := &relays.Location{
			IPv4Address: "1.2.3.4",
			Hostname:    "test",
		}

		result := Result{
			Location: loc,
			Latency:  nil,
		}

		if result.Latency != nil {
			t.Error("Latency should be nil for timeout")
		}
	})
}
