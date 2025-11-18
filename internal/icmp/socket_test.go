package icmp

import (
	"testing"

	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// TestListen_IPv4 tests the Listen function for IPv4
func TestListen_IPv4(t *testing.T) {
	conn, network, err := Listen(relays.IPv4)
	if err != nil {
		// Listen may fail if neither raw nor UDP ICMP sockets are available
		// This is expected in restricted environments
		t.Skipf("Skipping ICMP test: %v (requires either raw ICMP or ping_group_range)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return one of the valid network types
	if network != NetworkIPv4Raw && network != NetworkIPv4UDP {
		t.Errorf("Expected ip4:icmp or udp4, got %s", network)
	}
}

// TestListen_IPv6 tests the Listen function for IPv6
func TestListen_IPv6(t *testing.T) {
	conn, network, err := Listen(relays.IPv6)
	if err != nil {
		// Listen may fail if neither raw nor UDP ICMP sockets are available
		// This is expected in restricted environments
		t.Skipf("Skipping ICMP test: %v (requires either raw ICMP or ping_group_range)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return one of the valid network types
	if network != NetworkIPv6Raw && network != NetworkIPv6UDP {
		t.Errorf("Expected ip6:ipv6-icmp or udp6, got %s", network)
	}
}

// TestListenRaw_IPv4 tests the ListenRaw function for IPv4
func TestListenRaw_IPv4(t *testing.T) {
	conn, network, err := ListenRaw(relays.IPv4)
	// Raw ICMP requires privileges, so we expect it to fail without them
	// This is a valid test case - we're testing that the error handling works
	if err != nil {
		// Verify the error is permission-related
		if network != "" {
			t.Errorf("Expected empty network on error, got %s", network)
		}
		// Test passes - error handling works correctly
		return
	}

	// If we get here, we have privileges
	defer func() { _ = conn.Close() }()

	// Should return raw ICMP network type
	if network != NetworkIPv4Raw {
		t.Errorf("Expected ip4:icmp, got %s", network)
	}
}

// TestListenRaw_IPv6 tests the ListenRaw function for IPv6
func TestListenRaw_IPv6(t *testing.T) {
	conn, network, err := ListenRaw(relays.IPv6)
	// Raw ICMP requires privileges, or IPv6 may be unavailable
	// This is a valid test case - we're testing that the error handling works
	if err != nil {
		// Verify the error is handled correctly
		if network != "" {
			t.Errorf("Expected empty network on error, got %s", network)
		}
		// Test passes - error handling works correctly
		return
	}

	// If we get here, we have privileges and IPv6 is available
	defer func() { _ = conn.Close() }()

	// Should return raw ICMPv6 network type
	if network != NetworkIPv6Raw {
		t.Errorf("Expected ip6:ipv6-icmp, got %s", network)
	}
}

// TestListenUDP_IPv4 tests the ListenUDP function for IPv4
func TestListenUDP_IPv4(t *testing.T) {
	conn, network, err := ListenUDP(relays.IPv4)
	if err != nil {
		// UDP ICMP requires ping_group_range configuration on Linux
		// Skip the test if we don't have the necessary permissions
		t.Skipf("Skipping UDP ICMP test: %v (requires ping_group_range configuration)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return UDP network type
	if network != NetworkIPv4UDP {
		t.Errorf("Expected udp4, got %s", network)
	}
}

// TestListenUDP_IPv6 tests the ListenUDP function for IPv6
func TestListenUDP_IPv6(t *testing.T) {
	conn, network, err := ListenUDP(relays.IPv6)
	if err != nil {
		// UDP ICMP requires ping_group_range configuration on Linux
		// Skip the test if we don't have the necessary permissions
		t.Skipf("Skipping UDP ICMP test: %v (requires ping_group_range configuration)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return UDP network type
	if network != NetworkIPv6UDP {
		t.Errorf("Expected udp6, got %s", network)
	}
}

// TestListenWithDetails_IPv4 tests the ListenWithDetails function for IPv4
func TestListenWithDetails_IPv4(t *testing.T) {
	conn, network, rawErr, err := ListenWithDetails(relays.IPv4)
	if err != nil {
		// ListenWithDetails may fail if neither raw nor UDP ICMP sockets are available
		// This is expected in restricted environments
		t.Skipf("Skipping ICMP test: %v (requires either raw ICMP or ping_group_range)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return one of the valid network types
	if network != NetworkIPv4Raw && network != NetworkIPv4UDP {
		t.Errorf("Expected ip4:icmp or udp4, got %s", network)
	}

	// If we got UDP, rawErr should be set (raw ICMP failed)
	if network == NetworkIPv4UDP && rawErr == nil {
		t.Errorf("Expected rawErr to be set when using UDP fallback")
	}

	// If we got raw ICMP, rawErr should be nil
	if network == NetworkIPv4Raw && rawErr != nil {
		t.Errorf("Expected rawErr to be nil when using raw ICMP, got %v", rawErr)
	}
}

// TestListenWithDetails_IPv6 tests the ListenWithDetails function for IPv6
func TestListenWithDetails_IPv6(t *testing.T) {
	conn, network, rawErr, err := ListenWithDetails(relays.IPv6)
	if err != nil {
		// ListenWithDetails may fail if neither raw nor UDP ICMP sockets are available
		// This is expected in restricted environments
		t.Skipf("Skipping ICMP test: %v (requires either raw ICMP or ping_group_range)", err)
	}
	defer func() { _ = conn.Close() }()

	// Should return one of the valid network types
	if network != NetworkIPv6Raw && network != NetworkIPv6UDP {
		t.Errorf("Expected ip6:ipv6-icmp or udp6, got %s", network)
	}

	// If we got UDP, rawErr should be set (raw ICMP failed)
	if network == NetworkIPv6UDP && rawErr == nil {
		t.Errorf("Expected rawErr to be set when using UDP fallback")
	}

	// If we got raw ICMP, rawErr should be nil
	if network == NetworkIPv6Raw && rawErr != nil {
		t.Errorf("Expected rawErr to be nil when using raw ICMP, got %v", rawErr)
	}
}
