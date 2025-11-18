// Package logging provides logging utilities and types for mullvad-compass.
package logging

import "fmt"

// LogLevel represents logging verbosity
type LogLevel int

// Log level constants from most verbose to least verbose
const (
	LogLevelDebug   LogLevel = iota // Debug level - most verbose
	LogLevelInfo                    // Info level - informational messages
	LogLevelWarning                 // Warning level - warning messages
	LogLevelError                   // Error level - error messages only
)

// ParseLogLevel parses a log level string
func ParseLogLevel(s string) (LogLevel, error) {
	switch s {
	case "debug":
		return LogLevelDebug, nil
	case "info":
		return LogLevelInfo, nil
	case "warning":
		return LogLevelWarning, nil
	case "error":
		return LogLevelError, nil
	default:
		return LogLevelError, fmt.Errorf("invalid log level: %s (must be debug, info, warning, or error)", s)
	}
}
