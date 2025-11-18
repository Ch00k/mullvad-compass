package main

import (
	"testing"
	"time"
)

func TestPingLocations(t *testing.T) {
	t.Run("Empty locations", func(t *testing.T) {
		result, err := PingLocations([]Location{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected empty result for empty input, got %d locations", len(result))
		}
	})

	t.Run("Single location", func(t *testing.T) {
		locations := []Location{
			{
				IPv4Address: "127.0.0.1",
				Hostname:    "localhost",
				Country:     "Local",
				City:        "Test",
			},
		}

		result, err := PingLocations(locations)
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
		locations := []Location{
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

		result, err := PingLocations(locations)
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
		locations := []Location{
			{
				IPv4Address: "256.256.256.256", // Invalid IP
				Hostname:    "invalid",
				Country:     "Invalid",
				City:        "Test",
			},
		}

		result, err := PingLocations(locations)
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
		locations := []Location{
			{
				IPv4Address: "192.0.2.1", // TEST-NET-1, should be unreachable
				Hostname:    "unreachable",
				Country:     "Test",
				City:        "Test",
			},
		}

		result, err := PingLocations(locations)
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
		locations := make([]Location, maxWorkers*2)
		for i := range locations {
			locations[i] = Location{
				IPv4Address: "127.0.0.1",
				Hostname:    "test",
				Country:     "Test",
				City:        "Test",
			}
		}

		start := time.Now()
		result, err := PingLocations(locations)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		duration := time.Since(start)

		if len(result) != len(locations) {
			t.Errorf("Expected %d results, got %d", len(locations), len(result))
		}

		// With concurrency, should complete faster than sequential pings
		// Even if some fail, total time should be reasonable
		maxExpectedDuration := pingTimeout * time.Duration(len(locations)/maxWorkers+2)
		if duration > maxExpectedDuration {
			t.Logf("Warning: Pinging took longer than expected: %v (max expected: %v)", duration, maxExpectedDuration)
		}
	})
}

func TestPingWorker(t *testing.T) {
	t.Run("Worker processes work channel", func(t *testing.T) {
		workChan := make(chan *Location, 2)
		resultChan := make(chan PingResult, 2)

		loc1 := &Location{
			IPv4Address: "127.0.0.1",
			Hostname:    "test1",
		}
		loc2 := &Location{
			IPv4Address: "127.0.0.2",
			Hostname:    "test2",
		}

		workChan <- loc1
		workChan <- loc2
		close(workChan)

		go pingWorker(workChan, resultChan)

		// Collect results with timeout
		timeout := time.After(5 * time.Second)
		results := make([]PingResult, 0, 2)

	collect:
		for {
			select {
			case result, ok := <-resultChan:
				if !ok {
					break collect
				}
				results = append(results, result)
				if len(results) == 2 {
					break collect
				}
			case <-timeout:
				t.Fatal("Timeout waiting for ping results")
			}
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		// Verify results reference original locations
		found1, found2 := false, false
		for _, result := range results {
			if result.Location == loc1 {
				found1 = true
			}
			if result.Location == loc2 {
				found2 = true
			}
		}

		if !found1 || !found2 {
			t.Error("Results should reference original location pointers")
		}
	})
}

func TestPing(t *testing.T) {
	t.Run("Ping localhost", func(t *testing.T) {
		// This test may fail if running without proper permissions
		// or if ICMP is blocked by firewall
		result := ping("127.0.0.1")

		// We can't assert result is non-nil because it requires privileges
		// Just verify the function completes without panic
		if result != nil {
			if *result < 0 {
				t.Error("Latency should not be negative")
			}
			// Localhost ping should be very fast (< 100ms in most cases)
			if *result > 100 {
				t.Logf("Warning: Localhost ping latency seems high: %.2f ms", *result)
			}
		} else {
			t.Logf("Ping returned nil (likely missing ICMP privileges or firewall blocking)")
		}
	})

	t.Run("Ping invalid IP", func(t *testing.T) {
		result := ping("256.256.256.256")
		if result != nil {
			t.Error("Expected nil for invalid IP address")
		}
	})

	t.Run("Ping unreachable IP", func(t *testing.T) {
		// Using TEST-NET-1 address that should be unreachable
		result := ping("192.0.2.1")
		// Should timeout and return nil
		if result != nil {
			t.Logf("Warning: Expected nil for unreachable IP, got latency: %.2f ms", *result)
		}
	})

	t.Run("Ping respects timeout", func(t *testing.T) {
		start := time.Now()
		result := ping("192.0.2.1") // Unreachable IP
		duration := time.Since(start)

		// Should timeout around pingTimeout (1 second)
		// Allow some overhead for processing
		maxDuration := pingTimeout + 500*time.Millisecond
		if duration > maxDuration {
			t.Errorf("Ping took too long: %v (expected ~%v)", duration, pingTimeout)
		}

		if result != nil {
			t.Logf("Warning: Expected timeout (nil), got result: %.2f ms", *result)
		}
	})
}

func TestPingResult(t *testing.T) {
	t.Run("PingResult structure", func(t *testing.T) {
		latency := 12.34
		loc := &Location{
			IPv4Address: "1.2.3.4",
			Hostname:    "test",
		}

		result := PingResult{
			Location: loc,
			Latency:  &latency,
		}

		if result.Location != loc {
			t.Error("PingResult should reference the original location")
		}

		if result.Latency == nil {
			t.Error("Latency should not be nil")
		}

		if *result.Latency != 12.34 {
			t.Errorf("Expected latency 12.34, got %f", *result.Latency)
		}
	})

	t.Run("PingResult with nil latency", func(t *testing.T) {
		loc := &Location{
			IPv4Address: "1.2.3.4",
			Hostname:    "test",
		}

		result := PingResult{
			Location: loc,
			Latency:  nil,
		}

		if result.Latency != nil {
			t.Error("Latency should be nil for timeout")
		}
	})
}
