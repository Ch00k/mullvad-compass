package ping

import (
	"context"
	"strings"
	"sync"
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

func TestPingWorker(t *testing.T) {
	t.Run("Worker processes work channel", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		workChan := make(chan *relays.Location, 2)
		resultChan := make(chan Result, 2)

		loc1 := &relays.Location{
			IPv4Address: "127.0.0.1",
			Hostname:    "test1",
		}
		loc2 := &relays.Location{
			IPv4Address: "127.0.0.2",
			Hostname:    "test2",
		}

		workChan <- loc1
		workChan <- loc2
		close(workChan)

		pingTimeout := 500 * time.Millisecond
		go pingWorker(context.Background(), workChan, resultChan, pingTimeout, mgr, relays.IPv4)

		// Collect results with timeout
		timeout := time.After(5 * time.Second)
		results := make([]Result, 0, 2)

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

func TestSocketManager(t *testing.T) {
	t.Run("Create IPv4 socket manager", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create IPv4 socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		if mgr.conn == nil {
			t.Error("Socket manager should have a connection")
		}
		if mgr.protocol != protocolICMP {
			t.Errorf("Expected protocol %d, got %d", protocolICMP, mgr.protocol)
		}
	})

	t.Run("Create IPv6 socket manager", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv6)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create IPv6 socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		if mgr.conn == nil {
			t.Error("Socket manager should have a connection")
		}
		if mgr.protocol != protocolICMPv6 {
			t.Errorf("Expected protocol %d, got %d", protocolICMPv6, mgr.protocol)
		}
	})

	t.Run("Allocate unique sequence numbers", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Allocate multiple sequence numbers
		seqs := make(map[int]bool)
		for i := 0; i < 100; i++ {
			seq := mgr.allocateSeq()
			if seqs[seq] {
				t.Errorf("Duplicate sequence number: %d", seq)
			}
			seqs[seq] = true
		}

		if len(seqs) != 100 {
			t.Errorf("Expected 100 unique sequence numbers, got %d", len(seqs))
		}
	})

	t.Run("Concurrent sequence allocation", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		var wg sync.WaitGroup
		seqChan := make(chan int, 100)

		// Spawn multiple goroutines allocating sequences concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					seqChan <- mgr.allocateSeq()
				}
			}()
		}

		wg.Wait()
		close(seqChan)

		// Verify all sequences are unique
		seqs := make(map[int]bool)
		for seq := range seqChan {
			if seqs[seq] {
				t.Errorf("Duplicate sequence number from concurrent allocation: %d", seq)
			}
			seqs[seq] = true
		}

		if len(seqs) != 100 {
			t.Errorf("Expected 100 unique sequence numbers, got %d", len(seqs))
		}
	})

	t.Run("Ping localhost using socket manager", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		timeout := 500 * time.Millisecond
		result := mgr.Ping(context.Background(), "127.0.0.1", timeout)

		// May be nil if ICMP is blocked
		if result != nil {
			if *result < 0 {
				t.Error("Latency should not be negative")
			}
			if *result > 100 {
				t.Logf("Warning: Localhost ping latency seems high: %.2f ms", *result)
			}
		} else {
			t.Logf("Ping returned nil (likely firewall blocking)")
		}
	})

	t.Run("Ping invalid IP using socket manager", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		timeout := 500 * time.Millisecond
		result := mgr.Ping(context.Background(), "256.256.256.256", timeout)
		if result != nil {
			t.Error("Expected nil for invalid IP address")
		}
	})

	t.Run("Ping unreachable IP respects timeout", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		timeout := 500 * time.Millisecond
		start := time.Now()
		result := mgr.Ping(context.Background(), "192.0.2.1", timeout)
		duration := time.Since(start)

		if result != nil {
			t.Logf("Warning: Expected nil for unreachable IP, got latency: %.2f ms", *result)
		}

		// Should timeout around the specified timeout
		maxDuration := timeout + 200*time.Millisecond
		if duration > maxDuration {
			t.Errorf("Ping took too long: %v (expected ~%v)", duration, timeout)
		}
	})

	t.Run("Concurrent pings using same socket manager", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create socket manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		var wg sync.WaitGroup
		timeout := 500 * time.Millisecond
		numPings := 20

		results := make([]*float64, numPings)
		for i := 0; i < numPings; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				results[idx] = mgr.Ping(context.Background(), "127.0.0.1", timeout)
			}(i)
		}

		wg.Wait()

		// Check that we got some results (at least some should succeed)
		successCount := 0
		for _, result := range results {
			if result != nil {
				successCount++
				if *result < 0 {
					t.Error("Latency should not be negative")
				}
			}
		}

		// We expect at least some to succeed if ICMP works
		if successCount == 0 {
			t.Logf("Warning: No successful pings (likely firewall or missing privileges)")
		}
	})

	t.Run("Socket manager cleanup", func(t *testing.T) {
		mgr, err := newSocketManager(relays.IPv4)
		skipIfNoPermissions(t, err)
		if err != nil {
			t.Fatalf("Cannot create socket manager: %v", err)
		}

		// Close should complete without error
		err = mgr.Close()
		if err != nil {
			t.Errorf("Close should not error: %v", err)
		}

		// Subsequent close should also not panic
		_ = mgr.Close()
	})
}
