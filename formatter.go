package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// FormatTable formats locations as a table string
func FormatTable(locations []Location) string {
	if len(locations) == 0 {
		return ""
	}

	// Sort by latency (nil values last), with stable tie-breakers
	sort.SliceStable(locations, func(i, j int) bool {
		li, lj := locations[i].Latency, locations[j].Latency

		// Primary: Latency (nil last)
		if li == nil && lj != nil {
			return false
		}
		if li != nil && lj == nil {
			return true
		}
		if li != nil && lj != nil && *li != *lj {
			return *li < *lj
		}

		// Tie-breaker: Distance
		di, dj := locations[i].DistanceFromMyLocation, locations[j].DistanceFromMyLocation
		if di != nil && dj != nil && *di != *dj {
			return *di < *dj
		}

		// Final tie-breaker: Country then City
		if locations[i].Country != locations[j].Country {
			return locations[i].Country < locations[j].Country
		}
		return locations[i].City < locations[j].City
	})

	// Build table data
	headers := []string{"Country", "City", "Type", "IP", "Hostname", "Distance (km)", "Latency (ms)"}
	rows := make([][]string, len(locations))

	for i, loc := range locations {
		rows[i] = []string{
			loc.Country,
			loc.City,
			loc.Type,
			loc.IPv4Address,
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
	output.WriteString(strings.Join(headerParts, "  "))
	output.WriteString("\n")

	// Separator row
	separators := make([]string, len(headers))
	for i, width := range widths {
		separators[i] = strings.Repeat("-", width)
	}
	output.WriteString(strings.Join(separators, "  "))
	output.WriteString("\n")

	// Data rows
	for _, row := range rows {
		rowParts := make([]string, len(row))
		for i, cell := range row {
			rowParts[i] = padRight(cell, widths[i])
		}
		output.WriteString(strings.Join(rowParts, "  "))
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
	return fmt.Sprintf("%.2f", *distance)
}

// formatLatency formats a latency value for display
func formatLatency(latency *float64) string {
	if latency == nil {
		return "timeout"
	}
	return fmt.Sprintf("%.2f", *latency)
}
