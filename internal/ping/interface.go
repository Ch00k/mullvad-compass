package ping

import (
	"context"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// Pinger is an interface for ICMP ping operations
type Pinger interface {
	// Ping sends an ICMP echo request and returns the latency in milliseconds
	// Returns nil if the ping times out, fails, or context is cancelled
	Ping(ctx context.Context, ipAddr string, timeout time.Duration) *float64

	// Close cleans up resources
	Close() error
}

// PingerFactory creates Pinger instances
type PingerFactory interface {
	// CreatePinger creates a new Pinger for the specified IP version
	CreatePinger(ipVersion relays.IPVersion) (Pinger, error)
}

// defaultPingerFactory is the production implementation
type defaultPingerFactory struct{}

// NewDefaultPingerFactory creates a new default pinger factory
func NewDefaultPingerFactory() PingerFactory {
	return &defaultPingerFactory{}
}

// CreatePinger creates a real socket manager
func (f *defaultPingerFactory) CreatePinger(ipVersion relays.IPVersion) (Pinger, error) {
	return newSocketManager(ipVersion)
}

// Ensure socketManager implements Pinger
var _ Pinger = (*socketManager)(nil)
