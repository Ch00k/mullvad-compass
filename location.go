package main

// Location represents a Mullvad VPN server location with its properties and measured metrics
type Location struct {
	IPv4Address            string
	Country                string
	Latitude               float64
	Longitude              float64
	Hostname               string
	Type                   string // "openvpn" or "wireguard"
	City                   string
	IsActive               bool
	IsMullvadOwned         bool
	Provider               string
	Latency                *float64 // nil indicates timeout or error
	DistanceFromMyLocation *float64
}
