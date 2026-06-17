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
	Locations map[string]LocationEntry `json:"locations"`
	WireGuard WireGuardSection         `json:"wireguard"`
	Bridge    BridgeSection            `json:"bridge"`
}

// LocationEntry represents a location in the locations map
type LocationEntry struct {
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// WireGuardSection represents the wireguard section of the relays file
type WireGuardSection struct {
	Relays []WireGuardRelay `json:"relays"`
}

// WireGuardRelay represents a single WireGuard relay
type WireGuardRelay struct {
	Hostname               string        `json:"hostname"`
	Active                 bool          `json:"active"`
	Owned                  bool          `json:"owned"`
	Location               string        `json:"location"` // key into File.Locations
	Provider               string        `json:"provider"`
	IPv4AddrIn             string        `json:"ipv4_addr_in"`
	IPv6AddrIn             string        `json:"ipv6_addr_in"`
	IncludeInCountry       bool          `json:"include_in_country"`
	PublicKey              string        `json:"public_key"`
	Daita                  bool          `json:"daita"`
	ShadowsocksExtraAddrIn []string      `json:"shadowsocks_extra_addr_in"`
	Features               RelayFeatures `json:"features"`
}

// RelayFeatures represents anti-censorship capabilities on a WireGuard relay.
// Each field is null when absent or a JSON object when present.
type RelayFeatures struct {
	Daita *json.RawMessage `json:"daita"`
	QUIC  *json.RawMessage `json:"quic"`
	LWO   *json.RawMessage `json:"lwo"`
}

// BridgeSection represents the bridge section of the relays file
type BridgeSection struct {
	Relays []BridgeRelay `json:"relays"`
}

// BridgeRelay represents a single bridge relay
type BridgeRelay struct {
	Hostname         string `json:"hostname"`
	Active           bool   `json:"active"`
	Owned            bool   `json:"owned"`
	Location         string `json:"location"`
	Provider         string `json:"provider"`
	IPv4AddrIn       string `json:"ipv4_addr_in"`
	IncludeInCountry bool   `json:"include_in_country"`
}

// GetRelaysFilePath returns the platform-specific path to relays.json
func GetRelaysFilePath() (string, error) {
	return GetRelaysFilePathWithLogLevel(logging.LogLevelError)
}

// GetRelaysFilePathWithLogLevel returns the platform-specific path to relays.json with logging support
func GetRelaysFilePathWithLogLevel(logLevel logging.LogLevel) (string, error) {
	var basePath string

	switch {
	case os.Getenv("MULLVAD_COMPASS_RELAYS_FILE") != "":
		basePath = os.Getenv("MULLVAD_COMPASS_RELAYS_FILE")
	case runtime.GOOS == "linux":
		basePath = filepath.Join("/var/cache/mullvad-vpn", "relays.json")
	case runtime.GOOS == "darwin":
		basePath = filepath.Join("/Library/Caches/mullvad-vpn", "relays.json")
	case runtime.GOOS == "windows":
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

	locationCount := len(relays.Locations)
	wgRelayCount := len(relays.WireGuard.Relays)
	bridgeRelayCount := len(relays.Bridge.Relays)

	if logLevel <= logging.LogLevelInfo {
		log.Printf("Parsed relays file: %d locations, %d WireGuard relays, %d bridge relays",
			locationCount, wgRelayCount, bridgeRelayCount)
	}

	return &relays, nil
}

// shouldIncludeWireGuardRelay determines if a WireGuard relay should be included based on filter criteria
func shouldIncludeWireGuardRelay(
	relay WireGuardRelay,
	antiCensorship AntiCensorship,
	daita bool,
	ipVersion IPVersion,
) bool {
	if !relay.Active {
		return false
	}
	if !relay.IncludeInCountry {
		return false
	}
	if daita && !relay.Daita {
		return false
	}
	if antiCensorship != ACNone && !matchesAntiCensorshipFeatures(relay, antiCensorship) {
		return false
	}
	if ipVersion.IsIPv6() && relay.IPv6AddrIn == "" {
		return false
	}
	if !ipVersion.IsIPv6() && relay.IPv4AddrIn == "" {
		return false
	}
	return true
}

// GetLocations extracts Location objects from the relays file, optionally filtered by anti-censorship, DAITA, and IPv6.
// Returns the locations and the count of relays skipped due to unresolvable location keys.
func GetLocations(
	file *File,
	antiCensorship AntiCensorship,
	daita bool,
	ipVersion IPVersion,
) ([]Location, int, error) {
	locations := make([]Location, 0, len(file.WireGuard.Relays))
	var skipped int

	for _, relay := range file.WireGuard.Relays {
		locEntry, ok := file.Locations[relay.Location]
		if !ok {
			skipped++
			continue
		}

		if !shouldIncludeWireGuardRelay(relay, antiCensorship, daita, ipVersion) {
			continue
		}

		loc := Location{
			IPv4Address:    relay.IPv4AddrIn,
			IPv6Address:    relay.IPv6AddrIn,
			Country:        locEntry.Country,
			Latitude:       locEntry.Latitude,
			Longitude:      locEntry.Longitude,
			Hostname:       relay.Hostname,
			Type:           "wireguard",
			City:           locEntry.City,
			IsActive:       relay.Active,
			IsMullvadOwned: relay.Owned,
			Provider:       relay.Provider,
		}

		locations = append(locations, loc)
	}

	return locations, skipped, nil
}

// matchesAntiCensorshipFeatures checks if a relay matches the specified anti-censorship protocol
func matchesAntiCensorshipFeatures(relay WireGuardRelay, ac AntiCensorship) bool {
	switch ac {
	case LWO:
		return relay.Features.LWO != nil
	case QUIC:
		return relay.Features.QUIC != nil
	case Shadowsocks:
		// features.shadowsocks is not yet populated by Mullvad;
		// fall back to checking shadowsocks_extra_addr_in
		return len(relay.ShadowsocksExtraAddrIn) > 0
	default:
		return false
	}
}
