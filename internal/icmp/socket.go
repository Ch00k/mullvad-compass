// Package icmp provides utilities for creating and managing unprivileged ICMP datagram sockets.
package icmp

import (
	"log"

	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
	"golang.org/x/net/icmp"
)

// Network type constants for unprivileged ICMP datagram sockets
const (
	NetworkIPv4 = "udp4"
	NetworkIPv6 = "udp6"
)

// Address constants for listening on all interfaces
const (
	addrIPv4All = "0.0.0.0"
	addrIPv6All = "::"
)

// ListenWithDetails creates an unprivileged ICMP datagram socket
// Returns the connection, network type, and error
func ListenWithDetails(ipVersion relays.IPVersion) (*icmp.PacketConn, string, error) {
	return ListenWithDetailsAndLogLevel(ipVersion, logging.LogLevelError)
}

// ListenWithDetailsAndLogLevel creates an unprivileged ICMP datagram socket with logging support
func ListenWithDetailsAndLogLevel(
	ipVersion relays.IPVersion,
	logLevel logging.LogLevel,
) (*icmp.PacketConn, string, error) {
	var network, addr string
	if ipVersion.IsIPv6() {
		network = NetworkIPv6
		addr = addrIPv6All
	} else {
		network = NetworkIPv4
		addr = addrIPv4All
	}

	if logLevel <= logging.LogLevelDebug {
		log.Printf("Attempting to create ICMP datagram socket (%s on %s)", network, addr)
	}

	c, err := icmp.ListenPacket(network, addr)
	if err != nil {
		if logLevel <= logging.LogLevelError {
			log.Printf("Failed to create ICMP socket: %v", err)
		}
		return nil, "", err
	}

	if logLevel <= logging.LogLevelDebug {
		log.Printf("Successfully created ICMP datagram socket")
	}
	return c, network, nil
}

// Listen creates an unprivileged ICMP datagram socket
func Listen(ipVersion relays.IPVersion) (*icmp.PacketConn, string, error) {
	return ListenWithDetails(ipVersion)
}
