//go:build windows

package ping

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// isICMPAvailable checks if ICMP is available by pinging localhost
func isICMPAvailable(t *testing.T) bool {
	mgr, err := newWindowsSocketManager(relays.IPv4)
	if err != nil {
		return false
	}
	defer mgr.Close()

	ctx := context.Background()
	timeout := 1 * time.Second
	latency := mgr.Ping(ctx, "127.0.0.1", timeout)
	return latency != nil
}

func TestWindowsSocketManager_CreateClose(t *testing.T) {
	tests := []struct {
		name      string
		ipVersion relays.IPVersion
	}{
		{"IPv4", relays.IPv4},
		{"IPv6", relays.IPv6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := newWindowsSocketManager(tt.ipVersion)
			if err != nil {
				t.Fatalf("Failed to create socket manager: %v", err)
			}
			defer mgr.Close()

			if mgr.handle == 0 {
				t.Error("Expected valid handle, got 0")
			}

			if mgr.ipVersion != tt.ipVersion {
				t.Errorf("Expected IP version %v, got %v", tt.ipVersion, mgr.ipVersion)
			}
		})
	}
}

func TestWindowsSocketManager_CloseIdempotent(t *testing.T) {
	mgr, err := newWindowsSocketManager(relays.IPv4)
	if err != nil {
		t.Fatalf("Failed to create socket manager: %v", err)
	}

	// Close once
	err = mgr.Close()
	if err != nil {
		t.Errorf("First close failed: %v", err)
	}

	// Close again should fail gracefully
	// Windows will return an error for invalid handle, which is expected
	_ = mgr.Close()
}

func TestWindowsSocketManager_PingIPv4(t *testing.T) {
	mgr, err := newWindowsSocketManager(relays.IPv4)
	if err != nil {
		t.Fatalf("Failed to create socket manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()
	timeout := 2 * time.Second

	// First test localhost to verify ICMP works
	localhostLatency := mgr.Ping(ctx, "127.0.0.1", timeout)
	if localhostLatency == nil {
		t.Skip("ICMP appears to be unavailable on this system (localhost ping failed)")
	}

	// Test with known reachable IPv4 address (Google DNS)
	latency := mgr.Ping(ctx, "8.8.8.8", timeout)
	if latency == nil {
		// External ping failed but localhost worked - likely firewall/Azure restriction
		t.Skip("Outbound ICMP appears to be blocked (GitHub Actions/Azure blocks ICMP by default)")
	} else {
		t.Logf("Ping to 8.8.8.8: %.2f ms", *latency)
		if *latency <= 0 {
			t.Errorf("Expected positive latency, got %.2f", *latency)
		}
		if *latency > float64(timeout.Milliseconds()) {
			t.Errorf("Expected latency less than timeout %v ms, got %.2f ms", timeout.Milliseconds(), *latency)
		}
	}
}

func TestWindowsSocketManager_PingIPv6(t *testing.T) {
	// Check if IPv6 is available
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Skip("Cannot check IPv6 availability")
	}

	hasIPv6 := false
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() == nil && !ipnet.IP.IsLoopback() {
				hasIPv6 = true
				break
			}
		}
	}

	if !hasIPv6 {
		t.Skip("IPv6 not available on this system")
	}

	mgr, err := newWindowsSocketManager(relays.IPv6)
	if err != nil {
		t.Fatalf("Failed to create socket manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()
	timeout := 2 * time.Second

	// Test with known reachable IPv6 address (Google DNS)
	latency := mgr.Ping(ctx, "2001:4860:4860::8888", timeout)
	if latency == nil {
		t.Skip("Ping to 2001:4860:4860::8888 failed (may not have IPv6 route)")
	} else {
		t.Logf("Ping to 2001:4860:4860::8888: %.2f ms", *latency)
		if *latency <= 0 {
			t.Errorf("Expected positive latency, got %.2f", *latency)
		}
		if *latency > float64(timeout.Milliseconds()) {
			t.Errorf("Expected latency less than timeout %v ms, got %.2f ms", timeout.Milliseconds(), *latency)
		}
	}
}

func TestWindowsSocketManager_PingTimeout(t *testing.T) {
	mgr, err := newWindowsSocketManager(relays.IPv4)
	if err != nil {
		t.Fatalf("Failed to create socket manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()
	timeout := 100 * time.Millisecond

	// Use a non-routable IP address that should timeout
	// 192.0.2.0/24 is TEST-NET-1, reserved for documentation
	start := time.Now()
	latency := mgr.Ping(ctx, "192.0.2.1", timeout)
	elapsed := time.Since(start)

	if latency != nil {
		t.Errorf("Expected timeout, got latency: %.2f ms", *latency)
	}

	// Verify that timeout was respected (with some tolerance)
	if elapsed < timeout || elapsed > timeout*2 {
		t.Logf("Warning: elapsed time %v outside expected range [%v, %v]", elapsed, timeout, timeout*2)
	}
}

func TestWindowsSocketManager_ContextCancellation(t *testing.T) {
	mgr, err := newWindowsSocketManager(relays.IPv4)
	if err != nil {
		t.Fatalf("Failed to create socket manager: %v", err)
	}
	defer mgr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	timeout := 5 * time.Second

	// Cancel immediately
	cancel()

	// Ping should return nil due to cancelled context
	// Use localhost since we don't actually need the ping to succeed
	latency := mgr.Ping(ctx, "127.0.0.1", timeout)
	if latency != nil {
		t.Errorf("Expected nil due to cancelled context, got latency: %.2f ms", *latency)
	}
}

func TestWindowsSocketManager_ContextCancellationDuringPing(t *testing.T) {
	mgr, err := newWindowsSocketManager(relays.IPv4)
	if err != nil {
		t.Fatalf("Failed to create socket manager: %v", err)
	}
	defer mgr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	timeout := 5 * time.Second

	// Use a non-routable IP that would normally take a while
	start := time.Now()
	latency := mgr.Ping(ctx, "192.0.2.1", timeout)
	elapsed := time.Since(start)

	if latency != nil {
		t.Errorf("Expected nil due to context timeout, got latency: %.2f ms", *latency)
	}

	// Context should cancel before the full timeout
	if elapsed > 200*time.Millisecond {
		t.Errorf("Context cancellation took too long: %v", elapsed)
	}
}

func TestWindowsSocketManager_InvalidIP(t *testing.T) {
	mgr, err := newWindowsSocketManager(relays.IPv4)
	if err != nil {
		t.Fatalf("Failed to create socket manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()
	timeout := 1 * time.Second

	tests := []string{
		"not-an-ip",
		"256.256.256.256",
		"",
		"::g", // Invalid IPv6
	}

	for _, invalidIP := range tests {
		t.Run(invalidIP, func(t *testing.T) {
			latency := mgr.Ping(ctx, invalidIP, timeout)
			if latency != nil {
				t.Errorf("Expected nil for invalid IP %q, got latency: %.2f ms", invalidIP, *latency)
			}
		})
	}
}

func TestWindowsSocketManager_ConcurrentPings(t *testing.T) {
	if !isICMPAvailable(t) {
		t.Skip("ICMP appears to be unavailable on this system")
	}

	mgr, err := newWindowsSocketManager(relays.IPv4)
	if err != nil {
		t.Fatalf("Failed to create socket manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()
	timeout := 2 * time.Second

	// First verify external pings work at all
	testPing := mgr.Ping(ctx, "8.8.8.8", timeout)
	if testPing == nil {
		t.Skip("Outbound ICMP appears to be blocked (GitHub Actions/Azure blocks ICMP by default)")
	}

	// Test concurrent pings to the same destination
	concurrency := 10
	results := make(chan *float64, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			latency := mgr.Ping(ctx, "8.8.8.8", timeout)
			results <- latency
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < concurrency; i++ {
		result := <-results
		if result != nil {
			successCount++
			t.Logf("Concurrent ping %d: %.2f ms", i+1, *result)
		}
	}

	if successCount == 0 {
		t.Error("Expected at least some successful pings, got none")
	}

	t.Logf("Successful concurrent pings: %d/%d", successCount, concurrency)
}
