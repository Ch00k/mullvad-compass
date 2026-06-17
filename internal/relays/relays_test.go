package relays

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRelaysFile(t *testing.T) {
	relays, err := ParseRelaysFile("../../testdata/relays.json")
	if err != nil {
		t.Fatalf("Failed to parse relays.json: %v", err)
	}

	if len(relays.Locations) == 0 {
		t.Error("No locations found in relays.json")
	}

	if len(relays.WireGuard.Relays) == 0 {
		t.Error("No WireGuard relays found in relays.json")
	}
}

func TestGetRelaysFilePathHonorsEnvOverride(t *testing.T) {
	t.Setenv("MULLVAD_COMPASS_RELAYS_FILE", "../../testdata/relays.json")

	path, err := GetRelaysFilePath()
	if err != nil {
		t.Fatalf("GetRelaysFilePath failed: %v", err)
	}

	if path != "../../testdata/relays.json" {
		t.Errorf("Expected env override path, got %s", path)
	}
}

func TestGetRelaysFilePathReportsMissingFile(t *testing.T) {
	t.Setenv("MULLVAD_COMPASS_RELAYS_FILE", filepath.Join(t.TempDir(), "absent.json"))

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)

	_, err := GetRelaysFilePath()
	if err == nil {
		t.Fatal("Expected error for missing relays file, got nil")
	}

	if !strings.Contains(logBuf.String(), "relays.json not found") {
		t.Errorf("Expected not-found log, got: %q", logBuf.String())
	}
}

func TestParseRelaysFileErrors(t *testing.T) {
	t.Run("Missing file", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		_, err := ParseRelaysFile(filepath.Join(t.TempDir(), "absent.json"))
		if err == nil {
			t.Fatal("Expected error for missing file, got nil")
		}

		if !strings.Contains(logBuf.String(), "Failed to read relays file") {
			t.Errorf("Expected read-failure log, got: %q", logBuf.String())
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.json")
		if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
			t.Fatal(err)
		}

		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		_, err := ParseRelaysFile(path)
		if err == nil {
			t.Fatal("Expected error for invalid JSON, got nil")
		}

		if !strings.Contains(logBuf.String(), "Failed to parse JSON") {
			t.Errorf("Expected parse-failure log, got: %q", logBuf.String())
		}
	})
}

func TestGetLocations(t *testing.T) {
	relays, err := ParseRelaysFile("../../testdata/relays.json")
	if err != nil {
		t.Fatalf("Failed to parse relays.json: %v", err)
	}

	t.Run("Returns only WireGuard servers", func(t *testing.T) {
		locations, _, err := GetLocations(relays, ACNone, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Error("Expected some locations, got none")
		}

		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Expected only wireguard servers, found: %s (type: %s)", loc.Hostname, loc.Type)
			}
		}
	})

	t.Run("Verify location fields", func(t *testing.T) {
		locations, _, err := GetLocations(relays, ACNone, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No locations to verify")
		}

		loc := locations[0]
		if loc.IPv4Address == "" {
			t.Error("IPv4Address is empty")
		}
		if loc.Country == "" {
			t.Error("Country is empty")
		}
		if loc.City == "" {
			t.Error("City is empty")
		}
		if loc.Hostname == "" {
			t.Error("Hostname is empty")
		}
		if loc.Type == "" {
			t.Error("Type is empty")
		}
	})

	t.Run("Exclude relays with include_in_country=false", func(t *testing.T) {
		testRelays := &File{
			Locations: map[string]LocationEntry{
				"tc-tst": {
					City:      "Test City",
					Country:   "Test Country",
					Latitude:  50.0,
					Longitude: 10.0,
				},
			},
			WireGuard: WireGuardSection{
				Relays: []WireGuardRelay{
					{
						Hostname:         "included-server",
						IPv4AddrIn:       "1.1.1.1",
						Active:           true,
						Owned:            true,
						Provider:         "test",
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "test123",
					},
					{
						Hostname:         "excluded-server",
						IPv4AddrIn:       "2.2.2.2",
						Active:           true,
						Owned:            true,
						Provider:         "test",
						IncludeInCountry: false,
						Location:         "tc-tst",
						PublicKey:        "test456",
					},
				},
			},
		}

		locations, _, err := GetLocations(testRelays, ACNone, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) != 1 {
			t.Errorf("Expected 1 location, got %d", len(locations))
		}

		if len(locations) > 0 && locations[0].Hostname != "included-server" {
			t.Errorf("Expected included-server, got %s", locations[0].Hostname)
		}

		for _, loc := range locations {
			if loc.Hostname == "excluded-server" {
				t.Error("Found excluded-server which should have been filtered out")
			}
		}
	})

	t.Run("Filter by DAITA", func(t *testing.T) {
		locations, _, err := GetLocations(relays, ACNone, true, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No DAITA servers found in test data")
		}

		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server with DAITA filter: %s", loc.Hostname)
			}
		}
	})

	t.Run("Filter by LWO", func(t *testing.T) {
		locations, _, err := GetLocations(relays, LWO, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No LWO servers found in test data")
		}

		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server with LWO filter: %s", loc.Hostname)
			}
		}
	})

	t.Run("Filter by QUIC", func(t *testing.T) {
		locations, _, err := GetLocations(relays, QUIC, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No QUIC servers found in test data")
		}

		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server with QUIC filter: %s", loc.Hostname)
			}
		}
	})

	t.Run("Filter by Shadowsocks", func(t *testing.T) {
		locations, _, err := GetLocations(relays, Shadowsocks, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) == 0 {
			t.Skip("No Shadowsocks servers found in test data")
		}

		for _, loc := range locations {
			if loc.Type != "wireguard" {
				t.Errorf("Found non-wireguard server with Shadowsocks filter: %s", loc.Hostname)
			}
		}
	})

	t.Run("Filter locations without IPv6 when useIPv6 is true", func(t *testing.T) {
		testRelays := &File{
			Locations: map[string]LocationEntry{
				"tc-tst": {
					City:      "Test City",
					Country:   "Test",
					Latitude:  50.0,
					Longitude: 10.0,
				},
			},
			WireGuard: WireGuardSection{
				Relays: []WireGuardRelay{
					{
						Hostname:         "has-ipv6",
						IPv4AddrIn:       "1.1.1.1",
						IPv6AddrIn:       "2001:db8::1",
						Active:           true,
						Owned:            true,
						Provider:         "test",
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "test123",
					},
					{
						Hostname:         "no-ipv6",
						IPv4AddrIn:       "2.2.2.2",
						IPv6AddrIn:       "",
						Active:           true,
						Owned:            true,
						Provider:         "test",
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "test456",
					},
				},
			},
		}

		locations, _, err := GetLocations(testRelays, ACNone, false, IPv6)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) != 1 {
			t.Errorf("Expected 1 location with IPv6, got %d", len(locations))
		}

		if len(locations) > 0 && locations[0].Hostname != "has-ipv6" {
			t.Errorf("Expected has-ipv6, got %s", locations[0].Hostname)
		}

		for _, loc := range locations {
			if loc.Hostname == "no-ipv6" {
				t.Error("Found no-ipv6 which should have been filtered out")
			}
		}
	})

	t.Run("Include locations without IPv6 when useIPv6 is false", func(t *testing.T) {
		testRelays := &File{
			Locations: map[string]LocationEntry{
				"tc-tst": {
					City:      "Test City",
					Country:   "Test",
					Latitude:  50.0,
					Longitude: 10.0,
				},
			},
			WireGuard: WireGuardSection{
				Relays: []WireGuardRelay{
					{
						Hostname:         "has-ipv6",
						IPv4AddrIn:       "1.1.1.1",
						IPv6AddrIn:       "2001:db8::1",
						Active:           true,
						Owned:            true,
						Provider:         "test",
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "test123",
					},
					{
						Hostname:         "no-ipv6",
						IPv4AddrIn:       "2.2.2.2",
						IPv6AddrIn:       "",
						Active:           true,
						Owned:            true,
						Provider:         "test",
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "test456",
					},
				},
			},
		}

		locations, _, err := GetLocations(testRelays, ACNone, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) != 2 {
			t.Errorf("Expected 2 locations, got %d", len(locations))
		}
	})

	t.Run("Skip relays with unresolvable location key", func(t *testing.T) {
		testRelays := &File{
			Locations: map[string]LocationEntry{
				"tc-tst": {
					City:      "Test City",
					Country:   "Test",
					Latitude:  50.0,
					Longitude: 10.0,
				},
			},
			WireGuard: WireGuardSection{
				Relays: []WireGuardRelay{
					{
						Hostname:         "valid-location",
						IPv4AddrIn:       "1.1.1.1",
						Active:           true,
						Owned:            true,
						Provider:         "test",
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "test123",
					},
					{
						Hostname:         "bad-location",
						IPv4AddrIn:       "2.2.2.2",
						Active:           true,
						Owned:            true,
						Provider:         "test",
						IncludeInCountry: true,
						Location:         "nonexistent",
						PublicKey:        "test456",
					},
				},
			},
		}

		locations, skipped, err := GetLocations(testRelays, ACNone, false, IPv4)
		if err != nil {
			t.Fatalf("GetLocations failed: %v", err)
		}

		if len(locations) != 1 {
			t.Errorf("Expected 1 location, got %d", len(locations))
		}
		if skipped != 1 {
			t.Errorf("Expected 1 skipped relay, got %d", skipped)
		}
	})

	t.Run("Anti-censorship feature filtering with inline data", func(t *testing.T) {
		lwoObj := json.RawMessage(`{}`)
		quicObj := json.RawMessage(`{"addr_in":["1.2.3.4"]}`)

		testRelays := &File{
			Locations: map[string]LocationEntry{
				"tc-tst": {City: "Test", Country: "Test", Latitude: 50.0, Longitude: 10.0},
			},
			WireGuard: WireGuardSection{
				Relays: []WireGuardRelay{
					{
						Hostname:         "lwo-server",
						IPv4AddrIn:       "1.1.1.1",
						Active:           true,
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "k1",
						Features:         RelayFeatures{LWO: &lwoObj},
					},
					{
						Hostname:         "quic-server",
						IPv4AddrIn:       "2.2.2.2",
						Active:           true,
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "k2",
						Features:         RelayFeatures{QUIC: &quicObj},
					},
					{
						Hostname:               "ss-server",
						IPv4AddrIn:             "3.3.3.3",
						Active:                 true,
						IncludeInCountry:       true,
						Location:               "tc-tst",
						PublicKey:              "k3",
						ShadowsocksExtraAddrIn: []string{"4.4.4.4"},
					},
					{
						Hostname:         "plain-server",
						IPv4AddrIn:       "5.5.5.5",
						Active:           true,
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "k4",
					},
				},
			},
		}

		lwoLocs, _, _ := GetLocations(testRelays, LWO, false, IPv4)
		if len(lwoLocs) != 1 || lwoLocs[0].Hostname != "lwo-server" {
			t.Errorf("LWO filter: expected [lwo-server], got %v", hostnames(lwoLocs))
		}

		quicLocs, _, _ := GetLocations(testRelays, QUIC, false, IPv4)
		if len(quicLocs) != 1 || quicLocs[0].Hostname != "quic-server" {
			t.Errorf("QUIC filter: expected [quic-server], got %v", hostnames(quicLocs))
		}

		ssLocs, _, _ := GetLocations(testRelays, Shadowsocks, false, IPv4)
		if len(ssLocs) != 1 || ssLocs[0].Hostname != "ss-server" {
			t.Errorf("Shadowsocks filter: expected [ss-server], got %v", hostnames(ssLocs))
		}
	})

	t.Run("Filter by DAITA with inline data", func(t *testing.T) {
		testRelays := &File{
			Locations: map[string]LocationEntry{
				"tc-tst": {City: "Test", Country: "Test", Latitude: 50.0, Longitude: 10.0},
			},
			WireGuard: WireGuardSection{
				Relays: []WireGuardRelay{
					{
						Hostname:         "daita-server",
						IPv4AddrIn:       "1.1.1.1",
						Active:           true,
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "k1",
						Daita:            true,
					},
					{
						Hostname:         "plain-server",
						IPv4AddrIn:       "2.2.2.2",
						Active:           true,
						IncludeInCountry: true,
						Location:         "tc-tst",
						PublicKey:        "k2",
						Daita:            false,
					},
				},
			},
		}

		daitaLocs, _, _ := GetLocations(testRelays, ACNone, true, IPv4)
		if len(daitaLocs) != 1 || daitaLocs[0].Hostname != "daita-server" {
			t.Errorf("DAITA filter: expected [daita-server], got %v", hostnames(daitaLocs))
		}
	})
}

func TestMatchesAntiCensorshipFeaturesRejectsUnknown(t *testing.T) {
	if matchesAntiCensorshipFeatures(WireGuardRelay{}, ACNone) {
		t.Error("Expected ACNone to never match a relay's features")
	}
}

func hostnames(locs []Location) []string {
	names := make([]string, len(locs))
	for i, l := range locs {
		names[i] = l.Hostname
	}
	return names
}
