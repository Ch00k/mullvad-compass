package relays

import (
	"testing"
)

func TestAntiCensorshipString(t *testing.T) {
	tests := []struct {
		name string
		obf  AntiCensorship
		want string
	}{
		{"LWO", LWO, "lwo"},
		{"QUIC", QUIC, "quic"},
		{"Shadowsocks", Shadowsocks, "shadowsocks"},
		{"None", ACNone, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.obf.String(); got != tt.want {
				t.Errorf("AntiCensorship.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAntiCensorship(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    AntiCensorship
		wantErr bool
	}{
		{"lwo", "lwo", LWO, false},
		{"quic", "quic", QUIC, false},
		{"shadowsocks", "shadowsocks", Shadowsocks, false},
		{"empty", "", ACNone, false},
		{"invalid", "invalid", ACNone, true},
		{"uppercase", "LWO", ACNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAntiCensorship(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAntiCensorship() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAntiCensorship() = %v, want %v", got, tt.want)
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
