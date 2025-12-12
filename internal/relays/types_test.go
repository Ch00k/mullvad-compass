package relays

import (
	"testing"
)

func TestWireGuardObfuscationString(t *testing.T) {
	tests := []struct {
		name string
		obf  WireGuardObfuscation
		want string
	}{
		{"LWO", LWO, "lwo"},
		{"QUIC", QUIC, "quic"},
		{"Shadowsocks", Shadowsocks, "shadowsocks"},
		{"None", WGObfNone, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.obf.String(); got != tt.want {
				t.Errorf("WireGuardObfuscation.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseWireGuardObfuscation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    WireGuardObfuscation
		wantErr bool
	}{
		{"lwo", "lwo", LWO, false},
		{"quic", "quic", QUIC, false},
		{"shadowsocks", "shadowsocks", Shadowsocks, false},
		{"empty", "", WGObfNone, false},
		{"invalid", "invalid", WGObfNone, true},
		{"uppercase", "LWO", WGObfNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWireGuardObfuscation(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWireGuardObfuscation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseWireGuardObfuscation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version IPVersion
		want    string
	}{
		{"IPv4", IPv4, "ipv4"},
		{"IPv6", IPv6, "ipv6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("IPVersion.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPVersionIsIPv6(t *testing.T) {
	tests := []struct {
		name    string
		version IPVersion
		want    bool
	}{
		{"IPv4 is not IPv6", IPv4, false},
		{"IPv6 is IPv6", IPv6, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.IsIPv6(); got != tt.want {
				t.Errorf("IPVersion.IsIPv6() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerTypeString(t *testing.T) {
	tests := []struct {
		name       string
		serverType ServerType
		want       string
	}{
		{"WireGuard", WireGuard, "wireguard"},
		{"Bridge", Bridge, "bridge"},
		{"None", ServerTypeNone, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.serverType.String(); got != tt.want {
				t.Errorf("ServerType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseServerType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ServerType
		wantErr bool
	}{
		{"wireguard", "wireguard", WireGuard, false},
		{"bridge", "bridge", Bridge, false},
		{"empty", "", ServerTypeNone, false},
		{"invalid", "invalid", ServerTypeNone, true},
		{"uppercase", "WIREGUARD", ServerTypeNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseServerType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseServerType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseServerType() = %v, want %v", got, tt.want)
			}
		})
	}
}
