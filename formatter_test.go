package main

import (
	"strings"
	"testing"
)

func TestFormatTable(t *testing.T) {
	t.Run("Empty locations", func(t *testing.T) {
		result := FormatTable([]Location{})
		if result != "" {
			t.Errorf("Expected empty string for empty locations, got %q", result)
		}
	})

	t.Run("Single location with all fields", func(t *testing.T) {
		latency := 12.34
		distance := 123.45
		locations := []Location{
			{
				Country:                "Germany",
				City:                   "Berlin",
				Type:                   "wireguard",
				IPv4Address:            "185.65.134.1",
				Hostname:               "de-ber-wg-001",
				DistanceFromMyLocation: &distance,
				Latency:                &latency,
			},
		}

		result := FormatTable(locations)
		lines := strings.Split(strings.TrimSpace(result), "\n")

		if len(lines) != 3 {
			t.Errorf("Expected 3 lines (header, separator, data), got %d", len(lines))
		}

		// Check header
		if !strings.Contains(lines[0], "Country") {
			t.Error("Header should contain 'Country'")
		}
		if !strings.Contains(lines[0], "Latency") {
			t.Error("Header should contain 'Latency'")
		}

		// Check separator
		if !strings.Contains(lines[1], "---") {
			t.Error("Separator line should contain dashes")
		}

		// Check data
		if !strings.Contains(lines[2], "Germany") {
			t.Error("Data line should contain 'Germany'")
		}
		if !strings.Contains(lines[2], "Berlin") {
			t.Error("Data line should contain 'Berlin'")
		}
		if !strings.Contains(lines[2], "12.34") {
			t.Error("Data line should contain latency '12.34'")
		}
		if !strings.Contains(lines[2], "123.45") {
			t.Error("Data line should contain distance '123.45'")
		}
	})

	t.Run("Multiple locations sorted by latency", func(t *testing.T) {
		latency1 := 50.0
		latency2 := 10.0
		latency3 := 30.0
		distance := 100.0

		locations := []Location{
			{
				Country:                "Poland",
				City:                   "Warsaw",
				Type:                   "wireguard",
				IPv4Address:            "1.1.1.1",
				Hostname:               "pl-waw-wg-001",
				DistanceFromMyLocation: &distance,
				Latency:                &latency1, // 50ms - should be third
			},
			{
				Country:                "Germany",
				City:                   "Berlin",
				Type:                   "wireguard",
				IPv4Address:            "2.2.2.2",
				Hostname:               "de-ber-wg-001",
				DistanceFromMyLocation: &distance,
				Latency:                &latency2, // 10ms - should be first
			},
			{
				Country:                "France",
				City:                   "Paris",
				Type:                   "openvpn",
				IPv4Address:            "3.3.3.3",
				Hostname:               "fr-par-ovpn-001",
				DistanceFromMyLocation: &distance,
				Latency:                &latency3, // 30ms - should be second
			},
		}

		result := FormatTable(locations)
		lines := strings.Split(strings.TrimSpace(result), "\n")

		// Should have header + separator + 3 data lines
		if len(lines) != 5 {
			t.Errorf("Expected 5 lines, got %d", len(lines))
		}

		// Check ordering (skip header and separator)
		if !strings.Contains(lines[2], "Germany") {
			t.Error("First data line should be Germany (lowest latency)")
		}
		if !strings.Contains(lines[3], "France") {
			t.Error("Second data line should be France (middle latency)")
		}
		if !strings.Contains(lines[4], "Poland") {
			t.Error("Third data line should be Poland (highest latency)")
		}
	})

	t.Run("Locations with nil latency show timeout", func(t *testing.T) {
		distance := 100.0
		locations := []Location{
			{
				Country:                "Germany",
				City:                   "Berlin",
				Type:                   "wireguard",
				IPv4Address:            "1.1.1.1",
				Hostname:               "de-ber-wg-001",
				DistanceFromMyLocation: &distance,
				Latency:                nil,
			},
		}

		result := FormatTable(locations)
		if !strings.Contains(result, "timeout") {
			t.Error("Expected 'timeout' for nil latency")
		}
	})

	t.Run("Nil latency locations sorted last", func(t *testing.T) {
		latency := 10.0
		distance := 100.0

		locations := []Location{
			{
				Country:                "Poland",
				City:                   "Warsaw",
				Type:                   "wireguard",
				IPv4Address:            "1.1.1.1",
				Hostname:               "pl-waw-wg-001",
				DistanceFromMyLocation: &distance,
				Latency:                nil, // timeout - should be last
			},
			{
				Country:                "Germany",
				City:                   "Berlin",
				Type:                   "wireguard",
				IPv4Address:            "2.2.2.2",
				Hostname:               "de-ber-wg-001",
				DistanceFromMyLocation: &distance,
				Latency:                &latency, // 10ms - should be first
			},
		}

		result := FormatTable(locations)
		lines := strings.Split(strings.TrimSpace(result), "\n")

		// First data line should have latency value
		if !strings.Contains(lines[2], "Germany") {
			t.Error("First data line should be Germany (has latency)")
		}
		if !strings.Contains(lines[2], "10.00") {
			t.Error("First data line should show latency 10.00")
		}

		// Second data line should have timeout
		if !strings.Contains(lines[3], "Poland") {
			t.Error("Second data line should be Poland (timeout)")
		}
		if !strings.Contains(lines[3], "timeout") {
			t.Error("Second data line should show 'timeout'")
		}
	})

	t.Run("Table columns are properly aligned", func(t *testing.T) {
		latency := 12.34
		distance := 123.45
		locations := []Location{
			{
				Country:                "Germany",
				City:                   "Berlin",
				Type:                   "wireguard",
				IPv4Address:            "185.65.134.1",
				Hostname:               "de-ber-wg-001",
				DistanceFromMyLocation: &distance,
				Latency:                &latency,
			},
		}

		result := FormatTable(locations)
		lines := strings.Split(strings.TrimSpace(result), "\n")

		// All lines should have same structure (multiple spaces between columns)
		headerParts := strings.Fields(lines[0])
		dataParts := strings.Fields(lines[2])

		// Should have 7 columns
		if len(headerParts) != 7 {
			t.Errorf("Expected 7 header columns, got %d", len(headerParts))
		}
		if len(dataParts) != 7 {
			t.Errorf("Expected 7 data columns, got %d", len(dataParts))
		}
	})
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "String shorter than width",
			input:    "test",
			width:    10,
			expected: "test      ",
		},
		{
			name:     "String equal to width",
			input:    "test",
			width:    4,
			expected: "test",
		},
		{
			name:     "String longer than width",
			input:    "testing",
			width:    4,
			expected: "testing",
		},
		{
			name:     "Empty string",
			input:    "",
			width:    5,
			expected: "     ",
		},
		{
			name:     "Width zero",
			input:    "test",
			width:    0,
			expected: "test",
		},
		{
			name:     "String with unicode characters",
			input:    "Malmö",
			width:    10,
			expected: "Malmö     ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padRight(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("padRight(%q, %d) = %q, expected %q",
					tt.input, tt.width, result, tt.expected)
			}
			if len(result) < tt.width {
				t.Errorf("Result length %d is less than width %d",
					len(result), tt.width)
			}
		})
	}
}

func TestFormatDistance(t *testing.T) {
	tests := []struct {
		name     string
		distance *float64
		expected string
	}{
		{
			name:     "Nil distance",
			distance: nil,
			expected: "",
		},
		{
			name:     "Zero distance",
			distance: ptr(0.0),
			expected: "0.00",
		},
		{
			name:     "Small distance",
			distance: ptr(1.23),
			expected: "1.23",
		},
		{
			name:     "Large distance",
			distance: ptr(12345.67),
			expected: "12345.67",
		},
		{
			name:     "Distance with many decimals",
			distance: ptr(123.456789),
			expected: "123.46",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDistance(tt.distance)
			if result != tt.expected {
				t.Errorf("formatDistance(%v) = %q, expected %q",
					tt.distance, result, tt.expected)
			}
		})
	}
}

func TestFormatLatency(t *testing.T) {
	tests := []struct {
		name     string
		latency  *float64
		expected string
	}{
		{
			name:     "Nil latency (timeout)",
			latency:  nil,
			expected: "timeout",
		},
		{
			name:     "Zero latency",
			latency:  ptr(0.0),
			expected: "0.00",
		},
		{
			name:     "Small latency",
			latency:  ptr(1.23),
			expected: "1.23",
		},
		{
			name:     "Large latency",
			latency:  ptr(999.99),
			expected: "999.99",
		},
		{
			name:     "Latency with many decimals",
			latency:  ptr(12.3456789),
			expected: "12.35",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLatency(tt.latency)
			if result != tt.expected {
				t.Errorf("formatLatency(%v) = %q, expected %q",
					tt.latency, result, tt.expected)
			}
		})
	}
}

// Helper function to create pointer to float64
func ptr(f float64) *float64 {
	return &f
}
