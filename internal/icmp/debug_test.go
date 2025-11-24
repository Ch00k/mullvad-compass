//go:build windows

package icmp

import (
	"net"
	"testing"
	"time"
	"unsafe"
)

func TestDebugIPv4Ping(t *testing.T) {
	// Log structure sizes
	t.Logf("Structure sizes:")
	t.Logf("  IPOptionInformation: %d bytes", unsafe.Sizeof(IPOptionInformation{}))
	t.Logf("  IcmpEchoReply: %d bytes", unsafe.Sizeof(IcmpEchoReply{}))

	// Log field offsets
	t.Logf("IcmpEchoReply field offsets:")
	var reply IcmpEchoReply
	t.Logf("  Address: offset %d", unsafe.Offsetof(reply.Address))
	t.Logf("  Status: offset %d", unsafe.Offsetof(reply.Status))
	t.Logf("  RoundTripTime: offset %d", unsafe.Offsetof(reply.RoundTripTime))
	t.Logf("  DataSize: offset %d", unsafe.Offsetof(reply.DataSize))
	t.Logf("  Reserved: offset %d", unsafe.Offsetof(reply.Reserved))
	t.Logf("  Data: offset %d", unsafe.Offsetof(reply.Data))
	t.Logf("  Options: offset %d", unsafe.Offsetof(reply.Options))

	// Create handle
	handle, err := IcmpCreateFile()
	if err != nil {
		t.Fatalf("IcmpCreateFile failed: %v", err)
	}
	defer IcmpCloseHandle(handle)

	t.Logf("Created ICMP handle: %v", handle)

	// Test with 8.8.8.8
	ip := net.ParseIP("8.8.8.8")
	destAddr := IPv4ToUint32(ip)
	t.Logf("Destination address: %08x (IP: %s)", destAddr, ip)

	// Try localhost first to verify ICMP works at all
	t.Logf("\n=== Testing localhost (127.0.0.1) ===")
	localhostIP := net.ParseIP("127.0.0.1")
	localhostAddr := IPv4ToUint32(localhostIP)
	timeout := 1 * time.Second

	localhostReply, localhostErr := IcmpSendEcho(handle, localhostAddr, []byte("test"), timeout)
	if localhostErr != nil {
		t.Logf("Localhost ping FAILED: %v", localhostErr)
		t.Logf("This suggests ICMP is blocked or unavailable on this system")
	} else {
		t.Logf("Localhost ping succeeded: RTT=%dms", localhostReply.RoundTripTime)
	}

	// Send echo to 8.8.8.8
	t.Logf("\n=== Testing 8.8.8.8 ===")
	timeout = 2 * time.Second
	t.Logf("Sending ICMP echo with timeout %v", timeout)

	// Test with default data (nil, which will use "mullvad-compass")
	extReply, extErr := IcmpSendEcho(handle, destAddr, nil, timeout)

	if extErr != nil {
		t.Logf("IcmpSendEcho ERROR: %v", extErr)
		t.Logf("Error type: %T", extErr)

		// Try with explicit data
		t.Logf("\nRetrying with explicit data...")
		extReply2, extErr2 := IcmpSendEcho(handle, destAddr, []byte("test"), timeout)
		if extErr2 != nil {
			t.Logf("Second attempt also failed: %v", extErr2)
		} else {
			t.Logf("Second attempt succeeded!")
			extReply = extReply2
			extErr = nil
		}

		if extErr != nil {
			if localhostErr == nil {
				t.Logf("DIAGNOSIS: Localhost works but 8.8.8.8 fails")
				t.Logf("This indicates outbound ICMP is blocked (likely Azure/firewall policy)")
				t.Logf("GitHub Actions Windows runners run on Azure which blocks ICMP by default")
				t.Skip("Skipping test - ICMP blocked by infrastructure")
			}
			t.Logf("This error is the root cause of test failure")
			t.FailNow()
		}
	}

	t.Logf("Reply received:")
	t.Logf("  Address: %08x", extReply.Address)
	t.Logf("  Status: %d (%s)", extReply.Status, IPStatusToString(extReply.Status))
	t.Logf("  RoundTripTime: %d ms", extReply.RoundTripTime)
	t.Logf("  DataSize: %d", extReply.DataSize)

	if extReply.Status != IPSuccess {
		t.Errorf("Ping failed with status: %s", IPStatusToString(extReply.Status))
	}
}
