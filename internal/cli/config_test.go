package cli

import (
	"strings"
	"testing"

	"github.com/Ch00k/mullvad-compass/internal/relays"
)

func TestParseFlagsDefaults(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		cfg, err := ParseFlags([]string{}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.MaxDistance != 500.0 {
			t.Errorf("Expected maxDistance to be 500.0, got %f", cfg.MaxDistance)
		}
		if cfg.Timeout != 500 {
			t.Errorf("Expected timeout to be 500, got %d", cfg.Timeout)
		}
		if cfg.Workers != 25 {
			t.Errorf("Expected workers to be 25, got %d", cfg.Workers)
		}
		if cfg.ServerType != relays.ServerTypeNone {
			t.Errorf("Expected serverType to be relays.ServerTypeNone, got %v", cfg.ServerType)
		}
		if cfg.WireGuardObfuscation != relays.WGObfNone {
			t.Errorf("Expected wireGuardObfuscation to be relays.WGObfNone, got %v", cfg.WireGuardObfuscation)
		}
		if cfg.Daita {
			t.Error("Expected daita to be false, got true")
		}
		if cfg.IPVersion != relays.IPv4 {
			t.Errorf("Expected ipVersion to be relays.IPv4, got %v", cfg.IPVersion)
		}
		if cfg.ShowHelp {
			t.Error("Expected showHelp to be false, got true")
		}
		if cfg.ShowVersion {
			t.Error("Expected showVersion to be false, got true")
		}
	})
}

func TestParseFlagsHelp(t *testing.T) {
	t.Run("Help short flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-h"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.ShowHelp {
			t.Error("Expected showHelp to be true, got false")
		}
	})

	t.Run("Help long flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--help"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.ShowHelp {
			t.Error("Expected showHelp to be true, got false")
		}
	})
}

func TestParseFlagsVersion(t *testing.T) {
	t.Run("Version short flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-v"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.ShowVersion {
			t.Error("Expected showVersion to be true, got false")
		}
	})

	t.Run("Version long flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--version"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.ShowVersion {
			t.Error("Expected showVersion to be true, got false")
		}
	})
}

func TestParseFlagsServerType(t *testing.T) {
	t.Run("ServerType short flag with wireguard", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-s", "wireguard"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.ServerType != relays.WireGuard {
			t.Errorf("Expected serverType to be relays.WireGuard, got %v", cfg.ServerType)
		}
	})

	t.Run("ServerType long flag with openvpn", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--server-type", "openvpn"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.ServerType != relays.OpenVPN {
			t.Errorf("Expected serverType to be relays.OpenVPN, got %v", cfg.ServerType)
		}
	})

	t.Run("ServerType with invalid value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-s", "invalid"}, "dev")
		if err == nil {
			t.Error("Expected error for invalid server type, got nil")
		}
	})

	t.Run("ServerType without value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-s"}, "dev")
		if err == nil {
			t.Error("Expected error for missing server type value, got nil")
		}
	})
}

func TestParseFlagsWireguardObfuscation(t *testing.T) {
	t.Run("Obfuscation short flag with lwo", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-o", "lwo"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.WireGuardObfuscation != relays.LWO {
			t.Errorf("Expected wireGuardObfuscation to be relays.LWO, got %v", cfg.WireGuardObfuscation)
		}
	})

	t.Run("Obfuscation long flag with quic", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--wireguard-obfuscation", "quic"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.WireGuardObfuscation != relays.QUIC {
			t.Errorf("Expected wireGuardObfuscation to be relays.QUIC, got %v", cfg.WireGuardObfuscation)
		}
	})

	t.Run("Obfuscation with shadowsocks", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-o", "shadowsocks"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.WireGuardObfuscation != relays.Shadowsocks {
			t.Errorf("Expected wireGuardObfuscation to be relays.Shadowsocks, got %v", cfg.WireGuardObfuscation)
		}
	})

	t.Run("Obfuscation with invalid value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-o", "invalid"}, "dev")
		if err == nil {
			t.Error("Expected error for invalid obfuscation type, got nil")
		}
	})

	t.Run("Obfuscation without value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-o"}, "dev")
		if err == nil {
			t.Error("Expected error for missing obfuscation value, got nil")
		}
	})
}

func TestParseFlagsDAITA(t *testing.T) {
	t.Run("DAITA short flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-d"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.Daita {
			t.Error("Expected daita to be true, got false")
		}
	})

	t.Run("DAITA long flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--daita"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.Daita {
			t.Error("Expected daita to be true, got false")
		}
	})
}

func TestParseFlagsIPv6(t *testing.T) {
	t.Run("IPv6 short flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-6"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.IPVersion != relays.IPv6 {
			t.Errorf("Expected ipVersion to be relays.IPv6, got %v", cfg.IPVersion)
		}
	})

	t.Run("IPv6 long flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--ipv6"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.IPVersion != relays.IPv6 {
			t.Errorf("Expected ipVersion to be relays.IPv6, got %v", cfg.IPVersion)
		}
	})

	t.Run("IPv6 defaults to relays.IPv4", func(t *testing.T) {
		cfg, err := ParseFlags([]string{}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.IPVersion != relays.IPv4 {
			t.Errorf("Expected ipVersion to be relays.IPv4 by default, got %v", cfg.IPVersion)
		}
	})

	t.Run("IPv6 with other flags", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-6", "-s", "wireguard", "-m", "1000"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.IPVersion != relays.IPv6 {
			t.Errorf("Expected ipVersion to be relays.IPv6, got %v", cfg.IPVersion)
		}
		if cfg.ServerType != relays.WireGuard {
			t.Errorf("Expected serverType to be relays.WireGuard, got %v", cfg.ServerType)
		}
		if cfg.MaxDistance != 1000 {
			t.Errorf("Expected maxDistance to be 1000, got %f", cfg.MaxDistance)
		}
	})
}

func TestParseFlagsMaxDistance(t *testing.T) {
	t.Run("MaxDistance short flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-m", "250.5"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.MaxDistance != 250.5 {
			t.Errorf("Expected maxDistance to be 250.5, got %f", cfg.MaxDistance)
		}
	})

	t.Run("MaxDistance long flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--max-distance", "1000"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.MaxDistance != 1000 {
			t.Errorf("Expected maxDistance to be 1000, got %f", cfg.MaxDistance)
		}
	})

	t.Run("MaxDistance with invalid value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-m", "invalid"}, "dev")
		if err == nil {
			t.Error("Expected error for invalid max-distance value, got nil")
		}
	})

	t.Run("MaxDistance with negative value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-m", "-100"}, "dev")
		if err == nil {
			t.Error("Expected error for negative max-distance, got nil")
		}
	})

	t.Run("MaxDistance with zero value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-m", "0"}, "dev")
		if err == nil {
			t.Error("Expected error for zero max-distance, got nil")
		}
	})

	t.Run("MaxDistance without value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-m"}, "dev")
		if err == nil {
			t.Error("Expected error for missing max-distance value, got nil")
		}
	})

	t.Run("MaxDistance at maximum valid value", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-m", "20000"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.MaxDistance != 20000 {
			t.Errorf("Expected maxDistance to be 20000, got %f", cfg.MaxDistance)
		}
	})

	t.Run("MaxDistance above maximum", func(t *testing.T) {
		_, err := ParseFlags([]string{"-m", "20001"}, "dev")
		if err == nil {
			t.Error("Expected error for max-distance above 20000, got nil")
		}
		expectedError := "max-distance must be at most 20000 km"
		if err != nil && !strings.Contains(err.Error(), "max-distance must be at most 20000 km") {
			t.Errorf("Expected error containing %q, got: %v", expectedError, err)
		}
	})
}

func TestParseFlagsTimeout(t *testing.T) {
	t.Run("Timeout short flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-t", "1000"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.Timeout != 1000 {
			t.Errorf("Expected timeout to be 1000, got %d", cfg.Timeout)
		}
	})

	t.Run("Timeout long flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--timeout", "2000"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.Timeout != 2000 {
			t.Errorf("Expected timeout to be 2000, got %d", cfg.Timeout)
		}
	})

	t.Run("Timeout minimum valid value", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-t", "100"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.Timeout != 100 {
			t.Errorf("Expected timeout to be 100, got %d", cfg.Timeout)
		}
	})

	t.Run("Timeout maximum valid value", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-t", "5000"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.Timeout != 5000 {
			t.Errorf("Expected timeout to be 5000, got %d", cfg.Timeout)
		}
	})

	t.Run("Timeout below minimum", func(t *testing.T) {
		_, err := ParseFlags([]string{"-t", "50"}, "dev")
		if err == nil {
			t.Error("Expected error for timeout below minimum, got nil")
		}
	})

	t.Run("Timeout above maximum", func(t *testing.T) {
		_, err := ParseFlags([]string{"-t", "6000"}, "dev")
		if err == nil {
			t.Error("Expected error for timeout above maximum, got nil")
		}
	})

	t.Run("Timeout with invalid value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-t", "invalid"}, "dev")
		if err == nil {
			t.Error("Expected error for invalid timeout value, got nil")
		}
	})

	t.Run("Timeout without value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-t"}, "dev")
		if err == nil {
			t.Error("Expected error for missing timeout value, got nil")
		}
	})
}

func TestParseFlagsWorkers(t *testing.T) {
	t.Run("Workers short flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-w", "50"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.Workers != 50 {
			t.Errorf("Expected workers to be 50, got %d", cfg.Workers)
		}
	})

	t.Run("Workers long flag", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--workers", "100"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.Workers != 100 {
			t.Errorf("Expected workers to be 100, got %d", cfg.Workers)
		}
	})

	t.Run("Workers minimum valid value", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-w", "1"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.Workers != 1 {
			t.Errorf("Expected workers to be 1, got %d", cfg.Workers)
		}
	})

	t.Run("Workers maximum valid value", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-w", "200"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.Workers != 200 {
			t.Errorf("Expected workers to be 200, got %d", cfg.Workers)
		}
	})

	t.Run("Workers below minimum", func(t *testing.T) {
		_, err := ParseFlags([]string{"-w", "0"}, "dev")
		if err == nil {
			t.Error("Expected error for workers below minimum, got nil")
		}
	})

	t.Run("Workers above maximum", func(t *testing.T) {
		_, err := ParseFlags([]string{"-w", "250"}, "dev")
		if err == nil {
			t.Error("Expected error for workers above maximum, got nil")
		}
	})

	t.Run("Workers with invalid value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-w", "invalid"}, "dev")
		if err == nil {
			t.Error("Expected error for invalid workers value, got nil")
		}
	})

	t.Run("Workers without value", func(t *testing.T) {
		_, err := ParseFlags([]string{"-w"}, "dev")
		if err == nil {
			t.Error("Expected error for missing workers value, got nil")
		}
	})
}

func TestParseFlagsUnknownFlag(t *testing.T) {
	t.Run("Unknown short flag", func(t *testing.T) {
		_, err := ParseFlags([]string{"-x"}, "dev")
		if err == nil {
			t.Error("Expected error for unknown flag, got nil")
		}
	})

	t.Run("Unknown long flag", func(t *testing.T) {
		_, err := ParseFlags([]string{"--unknown"}, "dev")
		if err == nil {
			t.Error("Expected error for unknown flag, got nil")
		}
	})
}

func TestParseFlagsUnexpectedArgument(t *testing.T) {
	t.Run("Unexpected positional argument", func(t *testing.T) {
		_, err := ParseFlags([]string{"unexpected"}, "dev")
		if err == nil {
			t.Error("Expected error for unexpected argument, got nil")
		}
	})
}

func TestParseFlagesMultipleFlags(t *testing.T) {
	t.Run("Multiple flags combined", func(t *testing.T) {
		cfg, err := ParseFlags([]string{
			"-s", "wireguard",
			"-o", "lwo",
			"-d",
			"-6",
			"-m", "750.25",
			"-t", "1500",
			"-w", "50",
		}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.ServerType != relays.WireGuard {
			t.Errorf("Expected serverType to be relays.WireGuard, got %v", cfg.ServerType)
		}
		if cfg.WireGuardObfuscation != relays.LWO {
			t.Errorf("Expected wireGuardObfuscation to be relays.LWO, got %v", cfg.WireGuardObfuscation)
		}
		if !cfg.Daita {
			t.Error("Expected daita to be true, got false")
		}
		if cfg.IPVersion != relays.IPv6 {
			t.Errorf("Expected ipVersion to be relays.IPv6, got %v", cfg.IPVersion)
		}
		if cfg.MaxDistance != 750.25 {
			t.Errorf("Expected maxDistance to be 750.25, got %f", cfg.MaxDistance)
		}
		if cfg.Timeout != 1500 {
			t.Errorf("Expected timeout to be 1500, got %d", cfg.Timeout)
		}
		if cfg.Workers != 50 {
			t.Errorf("Expected workers to be 50, got %d", cfg.Workers)
		}
	})

	t.Run("Mix of short and long flags", func(t *testing.T) {
		cfg, err := ParseFlags([]string{
			"--server-type", "openvpn",
			"-m", "200",
			"--timeout", "1000",
			"-w", "10",
		}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.ServerType != relays.OpenVPN {
			t.Errorf("Expected serverType to be relays.OpenVPN, got %v", cfg.ServerType)
		}
		if cfg.MaxDistance != 200 {
			t.Errorf("Expected maxDistance to be 200, got %f", cfg.MaxDistance)
		}
		if cfg.Timeout != 1000 {
			t.Errorf("Expected timeout to be 1000, got %d", cfg.Timeout)
		}
		if cfg.Workers != 10 {
			t.Errorf("Expected workers to be 10, got %d", cfg.Workers)
		}
	})
}

func TestParseFlagsBestServerMode(t *testing.T) {
	t.Run("No args enables best server mode", func(t *testing.T) {
		cfg, err := ParseFlags([]string{}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.BestServerMode {
			t.Error("Expected BestServerMode to be true with no args")
		}
	})

	t.Run("Log level alone enables best server mode", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-l", "debug"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.BestServerMode {
			t.Error("Expected BestServerMode to be true with only log-level flag")
		}
	})

	t.Run("Long log level alone enables best server mode", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"--log-level", "info"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if !cfg.BestServerMode {
			t.Error("Expected BestServerMode to be true with only log-level flag")
		}
	})

	t.Run("Max distance disables best server mode", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-m", "1000"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.BestServerMode {
			t.Error("Expected BestServerMode to be false with max-distance flag")
		}
	})

	t.Run("Server type disables best server mode", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-s", "wireguard"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.BestServerMode {
			t.Error("Expected BestServerMode to be false with server-type flag")
		}
	})

	t.Run("Log level with max distance disables best server mode", func(t *testing.T) {
		cfg, err := ParseFlags([]string{"-l", "debug", "-m", "1000"}, "dev")
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}
		if cfg.BestServerMode {
			t.Error("Expected BestServerMode to be false with log-level and max-distance flags")
		}
	})
}
