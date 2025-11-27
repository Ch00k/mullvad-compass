// Package main provides the command-line interface for mullvad-compass.
package main

import "github.com/Ch00k/mullvad-compass/internal/relays"

// getDeterministicLocations returns a fixed set of server locations with predetermined values for testing/documentation
func getDeterministicLocations() []relays.Location {
	latency985 := 9.85
	latency1286 := 12.86
	latency1289 := 12.89
	latency1295 := 12.95
	latency1394 := 13.94
	latency1585 := 15.85
	latency1587 := 15.87
	latency1588a := 15.88
	latency1588b := 15.88
	latency1589 := 15.89
	latency1593 := 15.93
	latency1594 := 15.94
	latency1596 := 15.96
	latency1603 := 16.03
	distance15564 := 155.64
	distance23823 := 238.23

	return []relays.Location{
		{
			Country:                "Czech Republic",
			City:                   "Prague",
			Type:                   "wireguard",
			Hostname:               "cz-prg-wg-201",
			IPv4Address:            "178.249.209.162",
			DistanceFromMyLocation: &distance15564,
			Latency:                &latency985,
		},
		{
			Country:                "Czech Republic",
			City:                   "Prague",
			Type:                   "openvpn",
			Hostname:               "cz-prg-ovpn-102",
			IPv4Address:            "146.70.129.194",
			DistanceFromMyLocation: &distance15564,
			Latency:                &latency1286,
		},
		{
			Country:                "Czech Republic",
			City:                   "Prague",
			Type:                   "openvpn",
			Hostname:               "cz-prg-ovpn-101",
			IPv4Address:            "146.70.129.162",
			DistanceFromMyLocation: &distance15564,
			Latency:                &latency1289,
		},
		{
			Country:                "Czech Republic",
			City:                   "Prague",
			Type:                   "wireguard",
			Hostname:               "cz-prg-wg-202",
			IPv4Address:            "178.249.209.175",
			DistanceFromMyLocation: &distance15564,
			Latency:                &latency1295,
		},
		{
			Country:                "Czech Republic",
			City:                   "Prague",
			Type:                   "wireguard",
			Hostname:               "cz-prg-wg-102",
			IPv4Address:            "146.70.129.130",
			DistanceFromMyLocation: &distance15564,
			Latency:                &latency1394,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-001",
			IPv4Address:            "193.32.248.66",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1585,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-003",
			IPv4Address:            "193.32.248.68",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1587,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-008",
			IPv4Address:            "193.32.248.74",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1588a,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-007",
			IPv4Address:            "193.32.248.75",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1588b,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-004",
			IPv4Address:            "193.32.248.69",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1589,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-002",
			IPv4Address:            "193.32.248.67",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1593,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "openvpn",
			Hostname:               "de-ber-ovpn-001",
			IPv4Address:            "193.32.248.72",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1594,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-005",
			IPv4Address:            "193.32.248.70",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1596,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-006",
			IPv4Address:            "193.32.248.71",
			DistanceFromMyLocation: &distance23823,
			Latency:                &latency1603,
		},
	}
}
