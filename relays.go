package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// RelaysFile represents the structure of the relays.json file
type RelaysFile struct {
	Countries []Country `json:"countries"`
}

// Country represents a country in the relays file
type Country struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Cities []City `json:"cities"`
}

// City represents a city within a country
type City struct {
	Name      string  `json:"name"`
	Code      string  `json:"code"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Relays    []Relay `json:"relays"`
}

// Relay represents a single VPN server
type Relay struct {
	Hostname     string          `json:"hostname"`
	IPv4AddrIn   string          `json:"ipv4_addr_in"`
	Active       bool            `json:"active"`
	Owned        bool            `json:"owned"`
	Provider     string          `json:"provider"`
	EndpointData json.RawMessage `json:"endpoint_data"`
	Location     RelayLocation   `json:"location"`
}

// RelayLocation contains the geographic information for a relay
type RelayLocation struct {
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// WireguardEndpoint represents the wireguard endpoint data structure
type WireguardEndpoint struct {
	Wireguard struct {
		PublicKey string `json:"public_key"`
	} `json:"wireguard"`
}

// GetRelaysFilePath returns the platform-specific path to relays.json
func GetRelaysFilePath() (string, error) {
	var basePath string

	switch runtime.GOOS {
	case "linux":
		basePath = filepath.Join("/var/cache/mullvad-vpn", "relays.json")
	case "darwin":
		basePath = filepath.Join("/Library/Caches/mullvad-vpn", "relays.json")
	case "windows":
		basePath = filepath.Join("C:/ProgramData/Mullvad VPN/cache", "relays.json")
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return "", fmt.Errorf("relays.json not found at %s", basePath)
	}

	return basePath, nil
}

// ParseRelaysFile reads and parses the relays.json file
func ParseRelaysFile(path string) (*RelaysFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read relays file: %w", err)
	}

	var relays RelaysFile
	if err := json.Unmarshal(data, &relays); err != nil {
		return nil, fmt.Errorf("failed to parse relays file: %w", err)
	}

	return &relays, nil
}

// GetLocations extracts Location objects from the relays file, optionally filtered by server type
func GetLocations(relays *RelaysFile, serverType string) ([]Location, error) {
	var locations []Location

	for _, country := range relays.Countries {
		for _, city := range country.Cities {
			for _, relay := range city.Relays {
				// Determine the relay type
				relayType, err := determineRelayType(relay.EndpointData)
				if err != nil {
					continue // Skip relays we can't parse
				}

				// Filter by server type if specified
				if serverType != "" && relayType != serverType {
					continue
				}

				// Skip bridge servers
				if relayType == "bridge" {
					continue
				}

				loc := Location{
					IPv4Address:    relay.IPv4AddrIn,
					Country:        relay.Location.Country,
					Latitude:       relay.Location.Latitude,
					Longitude:      relay.Location.Longitude,
					Hostname:       relay.Hostname,
					Type:           relayType,
					City:           relay.Location.City,
					IsActive:       relay.Active,
					IsMullvadOwned: relay.Owned,
					Provider:       relay.Provider,
				}

				locations = append(locations, loc)
			}
		}
	}

	return locations, nil
}

// determineRelayType parses the endpoint_data field to determine the relay type
func determineRelayType(endpointData json.RawMessage) (string, error) {
	// Try to unmarshal as a string first (for "openvpn" or "bridge")
	var stringType string
	if err := json.Unmarshal(endpointData, &stringType); err == nil {
		return stringType, nil
	}

	// Try to unmarshal as an object (for wireguard)
	var objType WireguardEndpoint
	if err := json.Unmarshal(endpointData, &objType); err == nil {
		if objType.Wireguard.PublicKey != "" {
			return "wireguard", nil
		}
	}

	return "", fmt.Errorf("unknown endpoint_data format")
}
