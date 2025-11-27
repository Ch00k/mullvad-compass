// Package relays provides functions for parsing and filtering Mullvad relay servers.
package relays

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Ch00k/mullvad-compass/internal/logging"
)

// File represents the structure of the relays.json file
type File struct {
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
	Hostname         string          `json:"hostname"`
	IPv4AddrIn       string          `json:"ipv4_addr_in"`
	IPv6AddrIn       string          `json:"ipv6_addr_in"`
	Active           bool            `json:"active"`
	Owned            bool            `json:"owned"`
	Provider         string          `json:"provider"`
	IncludeInCountry bool            `json:"include_in_country"`
	EndpointData     json.RawMessage `json:"endpoint_data"`
	Location         RelayLocation   `json:"location"`
}

// RelayLocation contains the geographic information for a relay
type RelayLocation struct {
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// WireGuardEndpoint represents the wireguard endpoint data structure
type WireGuardEndpoint struct {
	Wireguard struct {
		PublicKey              string    `json:"public_key"`
		Daita                  bool      `json:"daita"`
		Lwo                    bool      `json:"lwo"`
		Quic                   *struct{} `json:"quic"`
		ShadowsocksExtraAddrIn []string  `json:"shadowsocks_extra_addr_in"`
	} `json:"wireguard"`
}

// GetRelaysFilePath returns the platform-specific path to relays.json
func GetRelaysFilePath() (string, error) {
	return GetRelaysFilePathWithLogLevel(logging.LogLevelError)
}

// GetRelaysFilePathWithLogLevel returns the platform-specific path to relays.json with logging support
func GetRelaysFilePathWithLogLevel(logLevel logging.LogLevel) (string, error) {
	var basePath string

	switch runtime.GOOS {
	case "linux":
		basePath = filepath.Join("/var/cache/mullvad-vpn", "relays.json")
	case "darwin":
		basePath = filepath.Join("/Library/Caches/mullvad-vpn", "relays.json")
	case "windows":
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		basePath = filepath.Join(programData, "Mullvad VPN", "cache", "relays.json")
	default:
		if logLevel <= logging.LogLevelError {
			log.Printf("Unsupported platform: %s", runtime.GOOS)
		}
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if logLevel <= logging.LogLevelDebug {
		log.Printf("Looking for relays.json at: %s", basePath)
	}

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if logLevel <= logging.LogLevelError {
			log.Printf("relays.json not found at %s", basePath)
		}
		return "", fmt.Errorf("relays.json not found at %s", basePath)
	}

	if logLevel <= logging.LogLevelDebug {
		log.Printf("Found relays.json at: %s", basePath)
	}

	return basePath, nil
}

// ParseRelaysFile reads and parses the relays.json file
func ParseRelaysFile(path string) (*File, error) {
	return ParseRelaysFileWithLogLevel(path, logging.LogLevelError)
}

// ParseRelaysFileWithLogLevel reads and parses the relays.json file with logging support
func ParseRelaysFileWithLogLevel(path string, logLevel logging.LogLevel) (*File, error) {
	if logLevel <= logging.LogLevelDebug {
		log.Printf("Reading relays file from: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if logLevel <= logging.LogLevelError {
			log.Printf("Failed to read relays file at %s: %v", path, err)
		}
		return nil, fmt.Errorf("failed to read relays file: %w", err)
	}

	if logLevel <= logging.LogLevelInfo {
		log.Printf("Read %d bytes from relays file", len(data))
	}

	var relays File
	if err := json.Unmarshal(data, &relays); err != nil {
		if logLevel <= logging.LogLevelError {
			log.Printf("Failed to parse JSON from relays file: %v", err)
		}
		return nil, fmt.Errorf("failed to parse relays file: %w", err)
	}

	countryCount := len(relays.Countries)
	var cityCount, relayCount int
	for _, country := range relays.Countries {
		cityCount += len(country.Cities)
		for _, city := range country.Cities {
			relayCount += len(city.Relays)
		}
	}

	if logLevel <= logging.LogLevelInfo {
		log.Printf("Parsed relays file: %d countries, %d cities, %d relays", countryCount, cityCount, relayCount)
	}

	return &relays, nil
}

// shouldIncludeRelay determines if a relay should be included based on filter criteria
func shouldIncludeRelay(
	relay Relay,
	relayType ServerType,
	serverType ServerType,
	wireGuardObfuscation WireGuardObfuscation,
	daita bool,
	ipVersion IPVersion,
) bool {
	// Skip inactive relays
	if !relay.Active {
		return false
	}

	// Skip relays excluded from country
	if !relay.IncludeInCountry {
		return false
	}

	// Skip bridge servers
	if relayType == Bridge {
		return false
	}

	// Filter by DAITA if specified
	if daita && !matchesDaita(relayType, relay.EndpointData) {
		return false
	}

	// Filter by wireguard obfuscation if specified
	if wireGuardObfuscation != WGObfNone &&
		!matchesWireGuardObfuscation(relayType, relay.EndpointData, wireGuardObfuscation) {
		return false
	}

	// Filter by server type if specified
	if serverType != ServerTypeNone && relayType != serverType {
		return false
	}

	// Filter by IP version
	if !matchesIPVersion(relay, ipVersion) {
		return false
	}

	return true
}

// matchesDaita checks if a relay matches DAITA requirements
func matchesDaita(relayType ServerType, endpointData json.RawMessage) bool {
	if relayType != WireGuard {
		return false
	}
	return hasDaita(endpointData)
}

// matchesWireGuardObfuscation checks if a relay matches wireguard obfuscation requirements
func matchesWireGuardObfuscation(
	relayType ServerType,
	endpointData json.RawMessage,
	obfuscationType WireGuardObfuscation,
) bool {
	if relayType != WireGuard {
		return false
	}
	return matchesObfuscation(endpointData, obfuscationType)
}

// matchesIPVersion checks if a relay has the required IP version
func matchesIPVersion(relay Relay, ipVersion IPVersion) bool {
	if ipVersion.IsIPv6() {
		return relay.IPv6AddrIn != ""
	}
	return relay.IPv4AddrIn != ""
}

// GetLocations extracts Location objects from the relays file, optionally filtered by server type, wireguard obfuscation, DAITA, and IPv6
// Returns the locations and the count of relays skipped due to unknown endpoint_data format
func GetLocations(
	relays *File,
	serverType ServerType,
	wireGuardObfuscation WireGuardObfuscation,
	daita bool,
	ipVersion IPVersion,
) ([]Location, int, error) {
	var locations []Location
	var skippedUnknownFormat int

	for _, country := range relays.Countries {
		for _, city := range country.Cities {
			for _, relay := range city.Relays {
				relayType, err := determineRelayType(relay.EndpointData)
				if err != nil {
					skippedUnknownFormat++
					continue
				}

				if !shouldIncludeRelay(relay, relayType, serverType, wireGuardObfuscation, daita, ipVersion) {
					continue
				}

				loc := Location{
					IPv4Address:    relay.IPv4AddrIn,
					IPv6Address:    relay.IPv6AddrIn,
					Country:        relay.Location.Country,
					Latitude:       relay.Location.Latitude,
					Longitude:      relay.Location.Longitude,
					Hostname:       relay.Hostname,
					Type:           relayType.String(),
					City:           relay.Location.City,
					IsActive:       relay.Active,
					IsMullvadOwned: relay.Owned,
					Provider:       relay.Provider,
				}

				locations = append(locations, loc)
			}
		}
	}

	return locations, skippedUnknownFormat, nil
}

// determineRelayType parses the endpoint_data field to determine the relay type
func determineRelayType(endpointData json.RawMessage) (ServerType, error) {
	if len(endpointData) == 0 {
		return ServerTypeNone, fmt.Errorf("empty endpoint_data")
	}

	// Check first non-whitespace byte to determine JSON type
	firstByte := endpointData[0]
	for i := 0; i < len(endpointData); i++ {
		if endpointData[i] != ' ' && endpointData[i] != '\t' && endpointData[i] != '\n' && endpointData[i] != '\r' {
			firstByte = endpointData[i]
			break
		}
	}

	switch firstByte {
	case '"':
		// String type (openvpn or bridge)
		var stringType string
		if err := json.Unmarshal(endpointData, &stringType); err != nil {
			return ServerTypeNone, fmt.Errorf("failed to unmarshal string endpoint_data: %w", err)
		}
		return ParseServerType(stringType)
	case '{':
		// Object type (wireguard)
		var objType WireGuardEndpoint
		if err := json.Unmarshal(endpointData, &objType); err != nil {
			return ServerTypeNone, fmt.Errorf("failed to unmarshal object endpoint_data: %w", err)
		}
		if objType.Wireguard.PublicKey != "" {
			return WireGuard, nil
		}
		return ServerTypeNone, fmt.Errorf("wireguard endpoint missing public key")
	}

	return ServerTypeNone, fmt.Errorf("unknown endpoint_data format")
}

// hasDaita checks if a wireguard endpoint has DAITA enabled
func hasDaita(endpointData json.RawMessage) bool {
	var endpoint WireGuardEndpoint
	if err := json.Unmarshal(endpointData, &endpoint); err != nil {
		return false
	}
	return endpoint.Wireguard.Daita
}

// matchesObfuscation checks if a wireguard endpoint matches the specified obfuscation type
func matchesObfuscation(endpointData json.RawMessage, obfuscationType WireGuardObfuscation) bool {
	var endpoint WireGuardEndpoint
	if err := json.Unmarshal(endpointData, &endpoint); err != nil {
		return false
	}

	switch obfuscationType {
	case LWO:
		return endpoint.Wireguard.Lwo
	case QUIC:
		return endpoint.Wireguard.Quic != nil
	case Shadowsocks:
		return len(endpoint.Wireguard.ShadowsocksExtraAddrIn) > 0
	default:
		return false
	}
}
