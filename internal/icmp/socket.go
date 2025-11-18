// Package icmp provides utilities for creating and managing ICMP sockets with automatic fallback between raw and UDP modes.
package icmp

import (
	"log"

	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
	"golang.org/x/net/icmp"
)

// SocketType represents the type of ICMP socket
type SocketType int

// ICMP socket type constants
const (
	SocketTypeRaw SocketType = iota // Raw ICMP socket requiring elevated privileges
	SocketTypeUDP                   // UDP ICMP socket available without special privileges
)

// Network type constants for ICMP sockets
const (
	NetworkIPv4Raw = "ip4:icmp"
	NetworkIPv6Raw = "ip6:ipv6-icmp"
	NetworkIPv4UDP = "udp4"
	NetworkIPv6UDP = "udp6"
)

// Address constants for listening on all interfaces
const (
	addrIPv4All = "0.0.0.0"
	addrIPv6All = "::"
)

// ListenWithDetails tries to create an ICMP connection, attempting raw ICMP first, then UDP fallback
// Returns the connection, network type, raw ICMP error (if it failed), and final error
func ListenWithDetails(ipVersion relays.IPVersion) (*icmp.PacketConn, string, error, error) {
	return ListenWithDetailsAndLogLevel(ipVersion, logging.LogLevelError)
}

// ListenWithDetailsAndLogLevel tries to create an ICMP connection with logging support
func ListenWithDetailsAndLogLevel(
	ipVersion relays.IPVersion,
	logLevel logging.LogLevel,
) (*icmp.PacketConn, string, error, error) {
	if ipVersion.IsIPv6() {
		// Try raw ICMPv6 first (requires privileges but gives more accurate results)
		if logLevel <= logging.LogLevelDebug {
			log.Printf("Attempting to create raw ICMPv6 socket (%s on %s)", NetworkIPv6Raw, addrIPv6All)
		}
		c, rawErr := icmp.ListenPacket(NetworkIPv6Raw, addrIPv6All)
		if rawErr == nil {
			if logLevel <= logging.LogLevelDebug {
				log.Printf("Successfully created raw ICMPv6 socket")
			}
			return c, NetworkIPv6Raw, nil, nil
		}

		if logLevel <= logging.LogLevelDebug {
			log.Printf("Failed to create raw ICMPv6 socket: %v", rawErr)
		}

		// Fallback to UDP datagram (unprivileged on macOS and Linux with net.ipv6.ping_group_range)
		if logLevel <= logging.LogLevelDebug {
			log.Printf("Attempting to create UDP ICMPv6 socket (%s on %s)", NetworkIPv6UDP, addrIPv6All)
		}
		c, err := icmp.ListenPacket(NetworkIPv6UDP, addrIPv6All)
		if err != nil {
			if logLevel <= logging.LogLevelError {
				log.Printf("Failed to create UDP ICMPv6 socket: %v", err)
			}
			return nil, "", rawErr, err
		}
		if logLevel <= logging.LogLevelDebug {
			log.Printf("Successfully created UDP ICMPv6 socket (fallback)")
		}
		return c, NetworkIPv6UDP, rawErr, nil
	}

	// Try raw ICMP first (requires privileges but gives more accurate results)
	if logLevel <= logging.LogLevelDebug {
		log.Printf("Attempting to create raw ICMP socket (%s on %s)", NetworkIPv4Raw, addrIPv4All)
	}
	c, rawErr := icmp.ListenPacket(NetworkIPv4Raw, addrIPv4All)
	if rawErr == nil {
		if logLevel <= logging.LogLevelDebug {
			log.Printf("Successfully created raw ICMP socket")
		}
		return c, NetworkIPv4Raw, nil, nil
	}

	if logLevel <= logging.LogLevelDebug {
		log.Printf("Failed to create raw ICMP socket: %v", rawErr)
	}

	// Fallback to UDP datagram (unprivileged on macOS and Linux with net.ipv4.ping_group_range)
	if logLevel <= logging.LogLevelDebug {
		log.Printf("Attempting to create UDP ICMP socket (%s on %s)", NetworkIPv4UDP, addrIPv4All)
	}
	c, err := icmp.ListenPacket(NetworkIPv4UDP, addrIPv4All)
	if err != nil {
		if logLevel <= logging.LogLevelError {
			log.Printf("Failed to create UDP ICMP socket: %v", err)
		}
		return nil, "", rawErr, err
	}
	if logLevel <= logging.LogLevelDebug {
		log.Printf("Successfully created UDP ICMP socket (fallback)")
	}
	return c, NetworkIPv4UDP, rawErr, nil
}

// Listen tries to create an ICMP connection, attempting raw ICMP first, then UDP fallback
func Listen(ipVersion relays.IPVersion) (*icmp.PacketConn, string, error) {
	conn, network, _, err := ListenWithDetails(ipVersion)
	return conn, network, err
}

// ListenRaw attempts to create a raw ICMP socket
func ListenRaw(ipVersion relays.IPVersion) (*icmp.PacketConn, string, error) {
	if ipVersion.IsIPv6() {
		c, err := icmp.ListenPacket(NetworkIPv6Raw, addrIPv6All)
		if err != nil {
			return nil, "", err
		}
		return c, NetworkIPv6Raw, nil
	}

	c, err := icmp.ListenPacket(NetworkIPv4Raw, addrIPv4All)
	if err != nil {
		return nil, "", err
	}
	return c, NetworkIPv4Raw, nil
}

// ListenUDP attempts to create a UDP ICMP socket
func ListenUDP(ipVersion relays.IPVersion) (*icmp.PacketConn, string, error) {
	if ipVersion.IsIPv6() {
		c, err := icmp.ListenPacket(NetworkIPv6UDP, addrIPv6All)
		if err != nil {
			return nil, "", err
		}
		return c, NetworkIPv6UDP, nil
	}

	c, err := icmp.ListenPacket(NetworkIPv4UDP, addrIPv4All)
	if err != nil {
		return nil, "", err
	}
	return c, NetworkIPv4UDP, nil
}
