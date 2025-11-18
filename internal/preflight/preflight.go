// Package preflight provides system capability checks for ICMP socket creation and IP version availability.
package preflight

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/Ch00k/mullvad-compass/internal/icmp"
	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// CheckResult contains the result of a preflight check
type CheckResult struct {
	Passed     bool
	Error      error
	SocketType icmp.SocketType
}

// CheckICMPPrivileges verifies the system can create ICMP sockets
// Also validates IP version availability (IPv4 or IPv6)
// Returns which socket type (raw or udp) succeeded for optimization
// Accepts logLevel to control logging of raw ICMP fallback
func CheckICMPPrivileges(ipVersion relays.IPVersion, logLevel logging.LogLevel) CheckResult {
	if logLevel <= logging.LogLevelDebug {
		log.Printf("Checking ICMP privileges for %s", ipVersion)
	}

	conn, network, rawErr, err := icmp.ListenWithDetailsAndLogLevel(ipVersion, logLevel)
	if err != nil {
		// Inspect error to provide specific guidance
		if strings.Contains(err.Error(), "permission denied") {
			if logLevel <= logging.LogLevelError {
				log.Printf("ICMP privilege check failed: insufficient privileges")
			}
			return CheckResult{
				Passed: false,
				Error:  fmt.Errorf("insufficient privileges to create ICMP socket\n%s", getPrivilegeHint()),
			}
		}
		// Check for IP version availability issues (works for both IPv4 and IPv6)
		if strings.Contains(err.Error(), "no suitable address") ||
			strings.Contains(err.Error(), "address family not supported") {
			if logLevel <= logging.LogLevelError {
				log.Printf("ICMP privilege check failed: %s not available on system", ipVersion)
			}
			return CheckResult{
				Passed: false,
				Error:  fmt.Errorf("%s not available on this system\nCheck network configuration", ipVersion),
			}
		}
		// Generic failure
		if logLevel <= logging.LogLevelError {
			log.Printf("ICMP privilege check failed: %v", err)
		}
		return CheckResult{
			Passed: false,
			Error:  fmt.Errorf("failed to create ICMP socket: %w", err),
		}
	}
	defer func() { _ = conn.Close() }()

	// Determine socket type from network string
	socketType := icmp.SocketTypeUDP
	if network == icmp.NetworkIPv4Raw || network == icmp.NetworkIPv6Raw {
		socketType = icmp.SocketTypeRaw
		if logLevel <= logging.LogLevelInfo {
			log.Printf("Using raw ICMP socket (%s)", network)
		}
	} else if logLevel <= logging.LogLevelWarning {
		// Raw ICMP failed, using UDP fallback - log if warning level or higher
		if rawErr != nil {
			log.Printf("Raw ICMP socket unavailable (%v), using UDP ICMP socket", rawErr)
		} else {
			log.Printf("Raw ICMP socket unavailable, using UDP ICMP socket")
		}
	}

	if logLevel <= logging.LogLevelDebug {
		log.Printf("ICMP privilege check passed, using %s socket", network)
	}

	return CheckResult{Passed: true, SocketType: socketType}
}

// getPrivilegeHint returns platform-specific guidance for gaining ICMP privileges
func getPrivilegeHint() string {
	switch runtime.GOOS {
	case "linux":
		return "Try running with sudo or grant capabilities: sudo setcap cap_net_raw+ep /path/to/mullvad-compass"
	case "darwin":
		return "Try running with sudo: sudo mullvad-compass"
	default:
		return "Try running with administrator/root privileges"
	}
}
