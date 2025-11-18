// Package formatter provides functionality for formatting and displaying server location data.
package formatter

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Ch00k/mullvad-compass/internal/api"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// SortLocationsByLatency sorts locations by latency (nil values last), with stable tie-breakers
func SortLocationsByLatency(locations []relays.Location) {
	slices.SortStableFunc(locations, func(a, b relays.Location) int {
		// Primary: Latency (nil last)
		if a.Latency == nil && b.Latency != nil {
			return 1
		}
		if a.Latency != nil && b.Latency == nil {
			return -1
		}
		if a.Latency != nil && b.Latency != nil {
			if c := cmp.Compare(*a.Latency, *b.Latency); c != 0 {
				return c
			}
		}

		// Tie-breaker: Distance
		if a.DistanceFromMyLocation != nil && b.DistanceFromMyLocation != nil {
			if c := cmp.Compare(*a.DistanceFromMyLocation, *b.DistanceFromMyLocation); c != 0 {
				return c
			}
		}

		// Final tie-breaker: Country then City
		if c := cmp.Compare(a.Country, b.Country); c != 0 {
			return c
		}
		return cmp.Compare(a.City, b.City)
	})
}

// FormatTable formats locations as a table string
func FormatTable(locations []relays.Location, useIPv6 bool) string {
	if len(locations) == 0 {
		return ""
	}

	// Build table data
	headers := []string{"Country", "City", "Type", "IP", "Hostname", "Distance (km)", "Latency (ms)"}
	rows := make([][]string, len(locations))

	for i, loc := range locations {
		ipAddr := loc.IPv4Address
		if useIPv6 {
			ipAddr = loc.IPv6Address
		}
		rows[i] = []string{
			loc.Country,
			loc.City,
			loc.Type,
			ipAddr,
			loc.Hostname,
			formatDistance(loc.DistanceFromMyLocation),
			formatLatency(loc.Latency),
		}
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = utf8.RuneCountInString(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			cellWidth := utf8.RuneCountInString(cell)
			if cellWidth > widths[i] {
				widths[i] = cellWidth
			}
		}
	}

	// Build table output
	var output strings.Builder

	// Header row
	headerParts := make([]string, len(headers))
	for i, header := range headers {
		headerParts[i] = padRight(header, widths[i])
	}
	output.WriteString(strings.Join(headerParts, "   "))
	output.WriteString("\n")

	// Separator row
	separators := make([]string, len(headers))
	for i, width := range widths {
		separators[i] = strings.Repeat("-", width)
	}
	output.WriteString(strings.Join(separators, "   "))
	output.WriteString("\n")

	// Data rows
	for _, row := range rows {
		rowParts := make([]string, len(row))
		for i, cell := range row {
			rowParts[i] = padRight(cell, widths[i])
		}
		output.WriteString(strings.Join(rowParts, "   "))
		output.WriteString("\n")
	}

	return output.String()
}

// padRight pads a string with spaces on the right to reach the specified width
func padRight(s string, width int) string {
	runeCount := utf8.RuneCountInString(s)
	if runeCount >= width {
		return s
	}
	return s + strings.Repeat(" ", width-runeCount)
}

// formatDistance formats a distance value for display
func formatDistance(distance *float64) string {
	if distance == nil {
		return ""
	}
	return fmt.Sprintf("%.0f", *distance)
}

// formatLatency formats a latency value for display
func formatLatency(latency *float64) string {
	if latency == nil {
		return "timeout"
	}
	return fmt.Sprintf("%.2f", *latency)
}

// FormatBestServer formats a single server in a compact format similar to --where-am-i
func FormatBestServer(loc relays.Location, useIPv6 bool) string {
	ipAddr := loc.IPv4Address
	if useIPv6 {
		ipAddr = loc.IPv6Address
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Country:    %s\n", loc.Country))
	output.WriteString(fmt.Sprintf("City:       %s\n", loc.City))
	output.WriteString(fmt.Sprintf("Distance:   %s km\n", formatDistance(loc.DistanceFromMyLocation)))
	output.WriteString(fmt.Sprintf("Hostname:   %s\n", loc.Hostname))
	output.WriteString(fmt.Sprintf("IP:         %s\n", ipAddr))
	output.WriteString(fmt.Sprintf("Latency:    %s ms\n", formatLatency(loc.Latency)))

	return output.String()
}

// FormatUserLocation formats user location information
func FormatUserLocation(loc api.UserLocation) string {
	mullvadStatus := "No"
	if loc.MullvadExitIP {
		mullvadStatus = "Yes"
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Country:                    %s\n", loc.Country))
	output.WriteString(fmt.Sprintf("City:                       %s\n", loc.City))
	output.WriteString(fmt.Sprintf("Latitude:                   %f\n", loc.Latitude))
	output.WriteString(fmt.Sprintf("Longitude:                  %f\n", loc.Longitude))
	output.WriteString(fmt.Sprintf("IP:                         %s\n", loc.IP))
	output.WriteString(fmt.Sprintf("Connected to Mullvad VPN:   %s\n", mullvadStatus))

	return output.String()
}
