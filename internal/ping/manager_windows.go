//go:build windows

package ping

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/icmp"
	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// windowsSocketManager manages Windows ICMP handle for IPv4 or IPv6
type windowsSocketManager struct {
	handle    icmp.Handle
	ipVersion relays.IPVersion
	closed    bool
	mu        sync.Mutex
}

// newWindowsSocketManager creates a new Windows socket manager for the given IP version
func newWindowsSocketManager(ipVersion relays.IPVersion) (*windowsSocketManager, error) {
	return newWindowsSocketManagerWithLogLevel(ipVersion, logging.LogLevelError)
}

// newWindowsSocketManagerWithLogLevel creates a new Windows socket manager with logging
func newWindowsSocketManagerWithLogLevel(ipVersion relays.IPVersion, logLevel logging.LogLevel) (*windowsSocketManager, error) {
	var handle icmp.Handle
	var err error

	if ipVersion.IsIPv6() {
		if logLevel <= logging.LogLevelDebug {
			log.Printf("Creating IPv6 ICMP handle")
		}
		handle, err = icmp.Icmp6CreateFile()
	} else {
		if logLevel <= logging.LogLevelDebug {
			log.Printf("Creating IPv4 ICMP handle")
		}
		handle, err = icmp.IcmpCreateFile()
	}

	if err != nil {
		if logLevel <= logging.LogLevelError {
			log.Printf("Failed to create ICMP handle: %v", err)
		}
		return nil, err
	}

	if logLevel <= logging.LogLevelDebug {
		log.Printf("Successfully created ICMP handle")
	}

	return &windowsSocketManager{
		handle:    handle,
		ipVersion: ipVersion,
	}, nil
}

// Ping sends an ICMP echo request and waits for a response
func (m *windowsSocketManager) Ping(ctx context.Context, ipAddr string, timeout time.Duration) *float64 {
	// Check context before starting
	if ctx.Err() != nil {
		return nil
	}

	// Parse IP address
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return nil
	}

	// Create channel for result
	resultChan := make(chan *float64, 1)

	// Run ping in goroutine to handle context cancellation
	go func() {
		var latencyMs *float64
		var err error

		if m.ipVersion.IsIPv6() {
			latencyMs, err = m.pingIPv6(ip, timeout)
		} else {
			latencyMs, err = m.pingIPv4(ip, timeout)
		}

		if err != nil {
			resultChan <- nil
			return
		}

		resultChan <- latencyMs
	}()

	// Wait for result or context cancellation
	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		// Note: Cannot cancel the syscall, but we can return early
		// The syscall will complete based on the timeout parameter
		return nil
	}
}

// pingIPv4 sends an IPv4 ICMP echo request
func (m *windowsSocketManager) pingIPv4(ip net.IP, timeout time.Duration) (*float64, error) {
	// Convert IP to uint32
	destAddr := icmp.IPv4ToUint32(ip)
	if destAddr == 0 {
		return nil, fmt.Errorf("invalid IPv4 address")
	}

	// Send echo request
	reply, err := icmp.IcmpSendEcho(m.handle, destAddr, nil, timeout)
	if err != nil {
		return nil, err
	}

	// Check reply status
	if reply.Status != icmp.IPSuccess {
		return nil, fmt.Errorf("ping failed: %s", icmp.IPStatusToString(reply.Status))
	}

	// Convert RTT to float64 milliseconds
	latencyMs := float64(reply.RoundTripTime)
	return &latencyMs, nil
}

// pingIPv6 sends an IPv6 ICMP echo request
func (m *windowsSocketManager) pingIPv6(ip net.IP, timeout time.Duration) (*float64, error) {
	// Ensure IP is in IPv6 format
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, fmt.Errorf("invalid IPv6 address")
	}

	// Send echo request
	reply, err := icmp.Icmp6SendEcho2(m.handle, ipv6, nil, timeout)
	if err != nil {
		return nil, err
	}

	// Check reply status
	if reply.Status != icmp.IPSuccess {
		return nil, fmt.Errorf("ping failed: %s", icmp.IPStatusToString(reply.Status))
	}

	// Convert RTT to float64 milliseconds
	latencyMs := float64(reply.RoundTripTime)
	return &latencyMs, nil
}

// Close cleans up the ICMP handle
func (m *windowsSocketManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already closed
	if m.closed {
		return nil
	}

	m.closed = true
	return icmp.IcmpCloseHandle(m.handle)
}

// Ensure windowsSocketManager implements Pinger
var _ Pinger = (*windowsSocketManager)(nil)
