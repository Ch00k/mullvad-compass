// Package main provides the command-line interface for mullvad-compass.
package main

import (
	"github.com/Ch00k/mullvad-compass/internal/api"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// getDeterministicUserLocation returns a fixed user location for testing/documentation
func getDeterministicUserLocation() api.UserLocation {
	return api.UserLocation{
		IP:            "203.0.113.42",
		Country:       "Germany",
		City:          "Dresden",
		Latitude:      51.0514,
		Longitude:     13.7341,
		MullvadExitIP: false,
	}
}

// getDeterministicLocations returns a fixed set of server locations with predetermined values for testing/documentation
func getDeterministicLocations() []relays.Location {
	latency978 := 9.78
	latency1301 := 13.01
	latency1394 := 13.94
	latency1586 := 15.86
	latency1588 := 15.88
	latency1589 := 15.89
	latency1591 := 15.91
	latency1593 := 15.93
	latency1595a := 15.95
	latency1595b := 15.95
	latency1599 := 15.99
	distance156 := 156.0
	distance238 := 238.0

	return []relays.Location{
		{
			Country:                "Czech Republic",
			City:                   "Prague",
			Type:                   "wireguard",
			Hostname:               "cz-prg-wg-201",
			IPv4Address:            "178.249.209.162",
			DistanceFromMyLocation: &distance156,
			Latency:                &latency978,
		},
		{
			Country:                "Czech Republic",
			City:                   "Prague",
			Type:                   "wireguard",
			Hostname:               "cz-prg-wg-202",
			IPv4Address:            "178.249.209.175",
			DistanceFromMyLocation: &distance156,
			Latency:                &latency1301,
		},
		{
			Country:                "Czech Republic",
			City:                   "Prague",
			Type:                   "wireguard",
			Hostname:               "cz-prg-wg-102",
			IPv4Address:            "146.70.129.130",
			DistanceFromMyLocation: &distance156,
			Latency:                &latency1394,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-007",
			IPv4Address:            "193.32.248.75",
			DistanceFromMyLocation: &distance238,
			Latency:                &latency1586,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-001",
			IPv4Address:            "193.32.248.66",
			DistanceFromMyLocation: &distance238,
			Latency:                &latency1588,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-005",
			IPv4Address:            "193.32.248.70",
			DistanceFromMyLocation: &distance238,
			Latency:                &latency1589,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-008",
			IPv4Address:            "193.32.248.74",
			DistanceFromMyLocation: &distance238,
			Latency:                &latency1591,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-003",
			IPv4Address:            "193.32.248.68",
			DistanceFromMyLocation: &distance238,
			Latency:                &latency1593,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-004",
			IPv4Address:            "193.32.248.69",
			DistanceFromMyLocation: &distance238,
			Latency:                &latency1595a,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-006",
			IPv4Address:            "193.32.248.71",
			DistanceFromMyLocation: &distance238,
			Latency:                &latency1595b,
		},
		{
			Country:                "Germany",
			City:                   "Berlin",
			Type:                   "wireguard",
			Hostname:               "de-ber-wg-002",
			IPv4Address:            "193.32.248.67",
			DistanceFromMyLocation: &distance238,
			Latency:                &latency1599,
		},
	}
}
