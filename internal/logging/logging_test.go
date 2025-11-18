package logging

import "testing"

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
		wantErr  bool
	}{
		{
			name:     "debug level",
			input:    "debug",
			expected: LogLevelDebug,
			wantErr:  false,
		},
		{
			name:     "info level",
			input:    "info",
			expected: LogLevelInfo,
			wantErr:  false,
		},
		{
			name:     "warning level",
			input:    "warning",
			expected: LogLevelWarning,
			wantErr:  false,
		},
		{
			name:     "error level",
			input:    "error",
			expected: LogLevelError,
			wantErr:  false,
		},
		{
			name:     "invalid level",
			input:    "invalid",
			expected: LogLevelError,
			wantErr:  true,
		},
		{
			name:     "uppercase not supported",
			input:    "DEBUG",
			expected: LogLevelError,
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: LogLevelError,
			wantErr:  true,
		},
		{
			name:     "mixed case not supported",
			input:    "Debug",
			expected: LogLevelError,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel_ErrorMessage(t *testing.T) {
	_, err := ParseLogLevel("invalid")
	if err == nil {
		t.Fatal("Expected error for invalid log level, got nil")
	}

	expectedMsg := "invalid log level: invalid (must be debug, info, warning, or error)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestLogLevelConstants(t *testing.T) {
	// Verify that log levels are ordered correctly (lower value = more verbose)
	if LogLevelDebug >= LogLevelInfo {
		t.Error("LogLevelDebug should be less than LogLevelInfo")
	}
	if LogLevelInfo >= LogLevelWarning {
		t.Error("LogLevelInfo should be less than LogLevelWarning")
	}
	if LogLevelWarning >= LogLevelError {
		t.Error("LogLevelWarning should be less than LogLevelError")
	}
}

func TestLogLevelComparison(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		compare  LogLevel
		expected bool
	}{
		{
			name:     "debug <= debug is true",
			level:    LogLevelDebug,
			compare:  LogLevelDebug,
			expected: true,
		},
		{
			name:     "debug <= info is true",
			level:    LogLevelDebug,
			compare:  LogLevelInfo,
			expected: true,
		},
		{
			name:     "info <= debug is false",
			level:    LogLevelInfo,
			compare:  LogLevelDebug,
			expected: false,
		},
		{
			name:     "error <= warning is false",
			level:    LogLevelError,
			compare:  LogLevelWarning,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level <= tt.compare
			if got != tt.expected {
				t.Errorf("%v <= %v = %v, want %v", tt.level, tt.compare, got, tt.expected)
			}
		})
	}
}
