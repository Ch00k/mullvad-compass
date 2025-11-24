//go:build !windows

package ping

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/relays"
)

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
			t.Error("Results should reference original locations")
		}
	})
}
