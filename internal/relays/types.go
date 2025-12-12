package relays

import "fmt"

// WireGuardObfuscation represents the obfuscation protocol for WireGuard connections.
type WireGuardObfuscation int

// WireGuard obfuscation protocol constants
const (
	WGObfNone   WireGuardObfuscation = iota // No obfuscation
	LWO                                     // Lightweight obfuscation
	QUIC                                    // QUIC protocol obfuscation
	Shadowsocks                             // Shadowsocks obfuscation
)

func (w WireGuardObfuscation) String() string {
	switch w {
	case LWO:
		return "lwo"
	case QUIC:
		return "quic"
	case Shadowsocks:
		return "shadowsocks"
	case WGObfNone:
		return ""
	default:
		return ""
	}
}

// ParseWireGuardObfuscation parses a WireGuard obfuscation string into its type.
func ParseWireGuardObfuscation(s string) (WireGuardObfuscation, error) {
	switch s {
	case "lwo":
		return LWO, nil
	case "quic":
		return QUIC, nil
	case "shadowsocks":
		return Shadowsocks, nil
	case "":
		return WGObfNone, nil
	default:
		return WGObfNone, fmt.Errorf("invalid wireguard obfuscation: %s (must be 'lwo', 'quic', or 'shadowsocks')", s)
	}
}

// IPVersion represents the IP protocol version (IPv4 or IPv6).
type IPVersion int

// IP version constants
const (
	IPv4 IPVersion = iota // IPv4 protocol
	IPv6                  // IPv6 protocol
)

func (v IPVersion) String() string {
	switch v {
	case IPv4:
		return "ipv4"
	case IPv6:
		return "ipv6"
	default:
		return "ipv4"
	}
}

// IsIPv6 returns true if the IP version is IPv6.
func (v IPVersion) IsIPv6() bool {
	return v == IPv6
}

// ServerType represents the type of VPN server.
type ServerType int

// Server type constants
const (
	ServerTypeNone ServerType = iota // No specific server type
	WireGuard                        // WireGuard server
	Bridge                           // Bridge server
)

func (s ServerType) String() string {
	switch s {
	case WireGuard:
		return "wireguard"
	case Bridge:
		return "bridge"
	case ServerTypeNone:
		return ""
	default:
		return ""
	}
}

// ParseServerType parses a server type string into its type.
func ParseServerType(s string) (ServerType, error) {
	switch s {
	case "wireguard":
		return WireGuard, nil
	case "bridge":
		return Bridge, nil
	case "":
		return ServerTypeNone, nil
	default:
		return ServerTypeNone, fmt.Errorf("unknown server type: %s", s)
	}
}

// Location represents a Mullvad VPN server location with its properties and measured metrics
type Location struct {
	IPv4Address            string
	IPv6Address            string
	Country                string
	Latitude               float64
	Longitude              float64
	Hostname               string
	Type                   string // "wireguard"
	City                   string
	IsActive               bool
	IsMullvadOwned         bool
	Provider               string
	Latency                *float64 // nil indicates timeout or error
	DistanceFromMyLocation *float64
}
