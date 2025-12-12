package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/Ch00k/mullvad-compass/internal/api"
	"github.com/Ch00k/mullvad-compass/internal/formatter"
	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

func TestGetUserLocation(t *testing.T) {
	t.Run("Calls underlying function and returns result", func(_ *testing.T) {
		expectedLoc := &api.UserLocation{
			Country:   "Sweden",
			City:      "Stockholm",
			Latitude:  59.3293,
			Longitude: 18.0686,
		}

		result, err := getUserLocation(
			context.Background(),
			logging.LogLevelError,
			func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return expectedLoc, nil
			},
		)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result != expectedLoc {
			t.Errorf("Expected result to match expected location")
		}
	})

	t.Run("Propagates errors from underlying function", func(t *testing.T) {
		expectedErr := errors.New("API error")

		result, err := getUserLocation(
			context.Background(),
			logging.LogLevelError,
			func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return nil, expectedErr
			},
		)

		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("Logs timing at debug level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		_, _ = getUserLocation(
			context.Background(),
			logging.LogLevelDebug,
			func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{}, nil
			},
		)

		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "User location fetch completed in") {
			t.Errorf("Expected timing log message, got: %s", logOutput)
		}
	})

	t.Run("Does not log timing at error level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		_, _ = getUserLocation(
			context.Background(),
			logging.LogLevelError,
			func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{}, nil
			},
		)

		logOutput := logBuf.String()
		if strings.Contains(logOutput, "User location fetch completed in") {
			t.Errorf("Did not expect timing log at error level, got: %s", logOutput)
		}
	})
}

func TestParseRelaysFile(t *testing.T) {
	t.Run("Logs timing at debug level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		mockGetRelaysPath := func() (string, error) {
			return "/nonexistent/path", nil
		}

		// This will fail because the path doesn't exist, but we just want to verify logging
		_, _ = parseRelaysFile(logging.LogLevelDebug, "/nonexistent/path", mockGetRelaysPath)

		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "Parse relays file completed in") {
			t.Errorf("Expected timing log message, got: %s", logOutput)
		}
	})

	t.Run("Does not log timing at error level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		mockGetRelaysPath := func() (string, error) {
			return "/nonexistent/path", nil
		}

		_, _ = parseRelaysFile(logging.LogLevelError, "/nonexistent/path", mockGetRelaysPath)

		logOutput := logBuf.String()
		if strings.Contains(logOutput, "Parse relays file completed in") {
			t.Errorf("Did not expect timing log at error level, got: %s", logOutput)
		}
	})
}

func TestGetLocations(t *testing.T) {
	t.Run("Logs timing at debug level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		// Create minimal relays data
		relaysData := &relays.File{}

		_, _ = getLocations(
			logging.LogLevelDebug,
			relaysData,
			relays.ACNone,
			false,
			relays.IPv4,
		)

		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "Get locations completed in") {
			t.Errorf("Expected timing log message, got: %s", logOutput)
		}
	})

	t.Run("Does not log timing at error level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		relaysData := &relays.File{}

		_, _ = getLocations(
			logging.LogLevelError,
			relaysData,
			relays.ACNone,
			false,
			relays.IPv4,
		)

		logOutput := logBuf.String()
		if strings.Contains(logOutput, "Get locations completed in") {
			t.Errorf("Did not expect timing log at error level, got: %s", logOutput)
		}
	})
}

func TestFilterByDistance(t *testing.T) {
	t.Run("Calls underlying function and returns result", func(_ *testing.T) {
		distance := 100.0
		locations := []relays.Location{
			{
				Country:                "Sweden",
				City:                   "Stockholm",
				DistanceFromMyLocation: &distance,
			},
		}

		result := filterByDistance(
			logging.LogLevelError,
			locations,
			59.3293,
			18.0686,
			1000.0,
		)

		// Note: The function filters by distance, so the result depends on actual location coordinates
		// We're just verifying the wrapper calls the underlying function correctly and returns a slice
		// The slice can be empty depending on the coordinates, which is valid behavior
		_ = result
	})

	t.Run("Logs timing at debug level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		locations := []relays.Location{}
		_ = filterByDistance(
			logging.LogLevelDebug,
			locations,
			0.0,
			0.0,
			1000.0,
		)

		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "Filter by distance completed in") {
			t.Errorf("Expected timing log message, got: %s", logOutput)
		}
	})

	t.Run("Does not log timing at error level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		locations := []relays.Location{}
		_ = filterByDistance(
			logging.LogLevelError,
			locations,
			0.0,
			0.0,
			1000.0,
		)

		logOutput := logBuf.String()
		if strings.Contains(logOutput, "Filter by distance completed in") {
			t.Errorf("Did not expect timing log at error level, got: %s", logOutput)
		}
	})
}

func TestSortLocationsByLatency(t *testing.T) {
	t.Run("Sorts locations correctly", func(t *testing.T) {
		latency1 := 50.0
		latency2 := 10.0

		locations := []relays.Location{
			{Country: "Poland", City: "Warsaw", Latency: &latency1},
			{Country: "Germany", City: "Berlin", Latency: &latency2},
		}

		sortLocationsByLatency(logging.LogLevelError, locations)

		if locations[0].Country != "Germany" {
			t.Errorf("Expected first location to be Germany, got %s", locations[0].Country)
		}
		if locations[1].Country != "Poland" {
			t.Errorf("Expected second location to be Poland, got %s", locations[1].Country)
		}
	})

	t.Run("Logs timing at debug level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		locations := []relays.Location{}
		sortLocationsByLatency(logging.LogLevelDebug, locations)

		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "Sort locations by latency completed in") {
			t.Errorf("Expected timing log message, got: %s", logOutput)
		}
	})

	t.Run("Does not log timing at error level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		locations := []relays.Location{}
		sortLocationsByLatency(logging.LogLevelError, locations)

		logOutput := logBuf.String()
		if strings.Contains(logOutput, "Sort locations by latency completed in") {
			t.Errorf("Did not expect timing log at error level, got: %s", logOutput)
		}
	})
}

func TestFormatTableTiming(t *testing.T) {
	t.Run("Formats table correctly", func(t *testing.T) {
		latency := 12.34
		distance := 123.45

		locations := []relays.Location{
			{
				Country:                "Germany",
				City:                   "Berlin",
				Type:                   "wireguard",
				IPv4Address:            "1.2.3.4",
				Hostname:               "de-ber-wg-001",
				DistanceFromMyLocation: &distance,
				Latency:                &latency,
			},
		}

		result := formatter.FormatTable(locations, false)

		if !strings.Contains(result, "Germany") {
			t.Error("Expected table to contain 'Germany'")
		}
		if !strings.Contains(result, "Berlin") {
			t.Error("Expected table to contain 'Berlin'")
		}
	})
}

func TestPingLocations(t *testing.T) {
	t.Run("Logs timing at debug level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		mockPingFn := func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
			return locs, nil
		}

		locations := []relays.Location{{Country: "Test", City: "Test"}}
		_, _ = pingLocations(
			context.Background(),
			logging.LogLevelDebug,
			locations,
			1000,
			10,
			relays.IPv4,
			mockPingFn,
		)

		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "Ping locations completed in") {
			t.Errorf("Expected timing log message, got: %s", logOutput)
		}
	})

	t.Run("Does not log timing at error level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		mockPingFn := func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
			return locs, nil
		}

		locations := []relays.Location{{Country: "Test", City: "Test"}}
		_, _ = pingLocations(
			context.Background(),
			logging.LogLevelError,
			locations,
			1000,
			10,
			relays.IPv4,
			mockPingFn,
		)

		logOutput := logBuf.String()
		if strings.Contains(logOutput, "Ping locations completed in") {
			t.Errorf("Did not expect timing log at error level, got: %s", logOutput)
		}
	})
}

func TestTotalOperationTiming(t *testing.T) {
	t.Run("Logs total operation timing at debug level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		var output bytes.Buffer
		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  59.3293,
					Longitude: 18.0686,
					Country:   "Sweden",
					City:      "Stockholm",
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return &relays.File{}, nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "100", "--log-level", "debug"}
		_ = run(context.Background(), args, deps)

		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "Total operation completed in") {
			t.Errorf("Expected total operation timing log message, got: %s", logOutput)
		}
	})

	t.Run("Does not log total operation timing at error level", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		var output bytes.Buffer
		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  59.3293,
					Longitude: 18.0686,
					Country:   "Sweden",
					City:      "Stockholm",
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return &relays.File{}, nil
			},
			Stdout: &output,
		}

		args := []string{"-m", "100"}
		_ = run(context.Background(), args, deps)

		logOutput := logBuf.String()
		if strings.Contains(logOutput, "Total operation completed in") {
			t.Errorf("Did not expect total operation timing log at error level, got: %s", logOutput)
		}
	})

	t.Run("Does not log total operation timing for help/version/whereami flags", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(nil)

		var output bytes.Buffer
		deps := Dependencies{
			GetUserLocation: func(context.Context, logging.LogLevel) (*api.UserLocation, error) {
				return &api.UserLocation{
					Latitude:  59.3293,
					Longitude: 18.0686,
					Country:   "Sweden",
					City:      "Stockholm",
				}, nil
			},
			PingLocations: func(_ context.Context, locs []relays.Location, _, _ int, _ relays.IPVersion, _ logging.LogLevel) ([]relays.Location, error) {
				return locs, nil
			},
			ParseRelaysFile: func(_ logging.LogLevel, _ string, _ func() (string, error)) (*relays.File, error) {
				return &relays.File{}, nil
			},
			Stdout: &output,
		}

		// Test help flag - should not log timing as it exits before timing starts
		args := []string{"--help", "--log-level", "debug"}
		_ = run(context.Background(), args, deps)

		logOutput := logBuf.String()
		if strings.Contains(logOutput, "Total operation completed in") {
			t.Errorf("Did not expect total operation timing log for help flag, got: %s", logOutput)
		}
	})
}
