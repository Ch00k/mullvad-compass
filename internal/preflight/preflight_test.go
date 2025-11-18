package preflight

import (
	"strings"
	"testing"

	"github.com/Ch00k/mullvad-compass/internal/icmp"
	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// TestCheckICMPPrivileges_IPv4_Success tests successful IPv4 ICMP check
// Note: This test requires appropriate privileges (CAP_NET_RAW or ping_group_range configured)
func TestCheckICMPPrivileges_IPv4_Success(t *testing.T) {
	result := CheckICMPPrivileges(relays.IPv4, logging.LogLevelError)
	if !result.Passed {
		// Skip if we don't have permissions - this test verifies the success path
		if result.Error != nil && strings.Contains(result.Error.Error(), "insufficient privileges") {
			t.Skipf("Skipping test due to insufficient permissions: %v", result.Error)
		}
		t.Fatalf("Expected check to pass, got error: %v", result.Error)
	}
	// Should return either Raw or UDP socket type
	if result.SocketType != icmp.SocketTypeRaw && result.SocketType != icmp.SocketTypeUDP {
		t.Errorf("Expected SocketTypeRaw or SocketTypeUDP, got %v", result.SocketType)
	}
}

// TestCheckICMPPrivileges_IPv6_Success tests successful IPv6 ICMP check
// Note: This test requires appropriate privileges and IPv6 availability
func TestCheckICMPPrivileges_IPv6_Success(t *testing.T) {
	result := CheckICMPPrivileges(relays.IPv6, logging.LogLevelError)
	if !result.Passed {
		// Skip if we don't have permissions - this test verifies the success path
		if result.Error != nil && strings.Contains(result.Error.Error(), "insufficient privileges") {
			t.Skipf("Skipping test due to insufficient permissions: %v", result.Error)
		}
		t.Fatalf("Expected check to pass, got error: %v", result.Error)
	}
	// Should return either Raw or UDP socket type
	if result.SocketType != icmp.SocketTypeRaw && result.SocketType != icmp.SocketTypeUDP {
		t.Errorf("Expected SocketTypeRaw or SocketTypeUDP, got %v", result.SocketType)
	}
}

// TestCheckICMPPrivileges_ErrorMessages tests that error messages are appropriate
func TestCheckICMPPrivileges_ErrorMessages(t *testing.T) {
	// We can't reliably trigger specific errors without complex setup,
	// so we just verify the error message format when a check fails
	result := CheckICMPPrivileges(relays.IPv4, logging.LogLevelError)
	if !result.Passed {
		if result.Error == nil {
			t.Fatalf("Expected error when check failed, got nil")
		}
		errMsg := result.Error.Error()
		// Error should contain useful information
		if len(errMsg) == 0 {
			t.Errorf("Expected non-empty error message")
		}
		// Should provide some guidance
		if strings.Contains(errMsg, "permission denied") {
			if !strings.Contains(errMsg, "sudo") && !strings.Contains(errMsg, "setcap") &&
				!strings.Contains(errMsg, "administrator") {
				t.Errorf("Permission error should include guidance, got: %s", errMsg)
			}
		}
	}
}

// TestGetPrivilegeHint tests that platform-specific hints are returned
func TestGetPrivilegeHint(t *testing.T) {
	hint := getPrivilegeHint()
	if len(hint) == 0 {
		t.Errorf("Expected non-empty privilege hint")
	}
	// Should mention either sudo, setcap, or administrator
	if !strings.Contains(hint, "sudo") && !strings.Contains(hint, "setcap") &&
		!strings.Contains(hint, "administrator") {
		t.Errorf("Expected hint to mention privilege escalation method, got: %s", hint)
	}
}
