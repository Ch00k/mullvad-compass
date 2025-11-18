package main

import (
	"fmt"
	"sort"
	"strings"
)

// FormatTable formats locations as a table string
func FormatTable(locations []Location) string {
	if len(locations) == 0 {
		return ""
	}

	// Sort by latency (nil values last)
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].Latency == nil {
			return false
		}
		if locations[j].Latency == nil {
			return true
		}
		return *locations[i].Latency < *locations[j].Latency
	})

	// Build table data
	headers := []string{"Country", "City", "Type", "IP", "Hostname", "Distance", "Latency"}
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
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
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
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
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
