//go:build !windows

package ping

import "github.com/Ch00k/mullvad-compass/internal/relays"

// createPlatformPinger creates a Unix-specific socket manager
func createPlatformPinger(ipVersion relays.IPVersion) (Pinger, error) {
	return newSocketManager(ipVersion)
}

// Ensure socketManager implements Pinger
var _ Pinger = (*socketManager)(nil)
