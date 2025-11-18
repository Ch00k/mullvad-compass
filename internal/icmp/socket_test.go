package icmp

import (
	"testing"

	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// TestListen_IPv4 tests the Listen function for IPv4
func TestListen_IPv4(t *testing.T) {
	conn, network, err := Listen(relays.IPv4)
	if err != nil {
		// Listen may fail if unprivileged ICMP sockets are not available
		// This is expected in restricted environments
		t.Skipf("Skipping ICMP test: %v (requires ping_group_range configuration)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return UDP network type for IPv4
	if network != NetworkIPv4 {
		t.Errorf("Expected udp4, got %s", network)
	}
}

// TestListen_IPv6 tests the Listen function for IPv6
func TestListen_IPv6(t *testing.T) {
	conn, network, err := Listen(relays.IPv6)
	if err != nil {
		// Listen may fail if unprivileged ICMP sockets are not available or IPv6 is unavailable
		// This is expected in restricted environments
		t.Skipf("Skipping ICMP test: %v (requires ping_group_range configuration and IPv6)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return UDP network type for IPv6
	if network != NetworkIPv6 {
		t.Errorf("Expected udp6, got %s", network)
	}
}

// TestListenWithDetails_IPv4 tests the ListenWithDetails function for IPv4
func TestListenWithDetails_IPv4(t *testing.T) {
	conn, network, err := ListenWithDetails(relays.IPv4)
	if err != nil {
		// ListenWithDetails may fail if unprivileged ICMP sockets are not available
		// This is expected in restricted environments
		t.Skipf("Skipping ICMP test: %v (requires ping_group_range configuration)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return UDP network type for IPv4
	if network != NetworkIPv4 {
		t.Errorf("Expected udp4, got %s", network)
	}
}

// TestListenWithDetails_IPv6 tests the ListenWithDetails function for IPv6
func TestListenWithDetails_IPv6(t *testing.T) {
	conn, network, err := ListenWithDetails(relays.IPv6)
	if err != nil {
		// ListenWithDetails may fail if unprivileged ICMP sockets are not available or IPv6 is unavailable
		// This is expected in restricted environments
		t.Skipf("Skipping ICMP test: %v (requires ping_group_range configuration and IPv6)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return UDP network type for IPv6
	if network != NetworkIPv6 {
		t.Errorf("Expected udp6, got %s", network)
	}
}
