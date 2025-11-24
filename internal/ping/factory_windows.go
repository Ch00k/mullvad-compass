//go:build windows

package ping

import "github.com/Ch00k/mullvad-compass/internal/relays"

// createPlatformPinger creates a Windows-specific socket manager
func createPlatformPinger(ipVersion relays.IPVersion) (Pinger, error) {
	return newWindowsSocketManager(ipVersion)
}
