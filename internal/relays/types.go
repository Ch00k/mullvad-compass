package relays

import "fmt"

// AntiCensorship represents the anti-censorship protocol for WireGuard connections.
type AntiCensorship int

// Anti-censorship protocol constants
const (
	ACNone      AntiCensorship = iota // No anti-censorship
	LWO                               // LWO protocol
	QUIC                              // QUIC protocol
	Shadowsocks                       // Shadowsocks protocol
)

func (w AntiCensorship) String() string {
	switch w {
	case LWO:
		return "lwo"
	case QUIC:
		return "quic"
	case Shadowsocks:
		return "shadowsocks"
	case ACNone:
		return ""
	default:
		return ""
	}
}

// ParseAntiCensorship parses an anti-censorship protocol string into its type.
func ParseAntiCensorship(s string) (AntiCensorship, error) {
	switch s {
	case "lwo":
		return LWO, nil
	case "quic":
		return QUIC, nil
	case "shadowsocks":
		return Shadowsocks, nil
	case "":
		return ACNone, nil
	default:
		return ACNone, fmt.Errorf("invalid anti-censorship protocol: %s (must be 'lwo', 'quic', or 'shadowsocks')", s)
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
