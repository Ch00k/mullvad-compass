// Package api provides functions for interacting with the Mullvad API.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/logging"
)

const (
	defaultAPIURL     = "https://am.i.mullvad.net/json"
	defaultTimeout    = 10 * time.Second
	defaultMaxRetries = 3
	defaultRetryDelay = 1 * time.Second
	defaultVersion    = "dev"
)

// Client encapsulates the HTTP client for interacting with the Mullvad API
type Client struct {
	httpClient *http.Client
	url        string
	maxRetries int
	retryDelay time.Duration
	version    string
	logLevel   logging.LogLevel
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithURL sets a custom API URL
func WithURL(url string) ClientOption {
	return func(c *Client) {
		c.url = url
	}
}

// WithTimeout sets a custom timeout for HTTP requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retry attempts
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryDelay sets the initial delay between retries
func WithRetryDelay(delay time.Duration) ClientOption {
	return func(c *Client) {
		c.retryDelay = delay
	}
}

// WithVersion sets the version string for the User-Agent header
func WithVersion(version string) ClientOption {
	return func(c *Client) {
		c.version = version
	}
}

// WithLogLevel sets the log level for the client
func WithLogLevel(logLevel logging.LogLevel) ClientOption {
	return func(c *Client) {
		c.logLevel = logLevel
	}
}

// NewClient creates a new API client with the given options
func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		url:        defaultAPIURL,
		maxRetries: defaultMaxRetries,
		retryDelay: defaultRetryDelay,
		version:    defaultVersion,
		logLevel:   logging.LogLevelError,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// UserLocation represents the response from Mullvad's location API
type UserLocation struct {
	IP            string  `json:"ip"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Country       string  `json:"country"`
	City          string  `json:"city"`
	MullvadExitIP bool    `json:"mullvad_exit_ip"`
}

// Error represents a structured error from the API client
type Error struct {
	StatusCode int
	Retriable  bool
	Err        error
}

func (e *Error) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("API error (status %d): %v", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("API error: %v", e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// isRetriableStatusCode determines if an HTTP status code should trigger a retry
func isRetriableStatusCode(statusCode int) bool {
	return statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusRequestTimeout ||
		statusCode >= 500
}

// GetUserLocation fetches the user's current geographic location from Mullvad API
func (c *Client) GetUserLocation(ctx context.Context) (*UserLocation, error) {
	var lastErr error

	if c.logLevel <= logging.LogLevelDebug {
		log.Printf("Fetching user location from %s (max retries: %d)", c.url, c.maxRetries)
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.retryDelay * time.Duration(1<<uint(attempt-1))
			if c.logLevel <= logging.LogLevelWarning {
				log.Printf("Retrying API request (attempt %d/%d) after %v delay", attempt+1, c.maxRetries+1, delay)
			}
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				if c.logLevel <= logging.LogLevelError {
					log.Printf("API request cancelled: %v", ctx.Err())
				}
				return nil, ctx.Err()
			}
		}

		location, err := c.doGetUserLocation(ctx)
		if err == nil {
			if c.logLevel <= logging.LogLevelInfo {
				log.Printf(
					"Successfully fetched user location: %s, %s (%.4f, %.4f)",
					location.City,
					location.Country,
					location.Latitude,
					location.Longitude,
				)
			}
			return location, nil
		}

		lastErr = err

		// Check if it's a structured APIError
		var apiErr *Error
		if errors.As(err, &apiErr) {
			if !apiErr.Retriable {
				if c.logLevel <= logging.LogLevelError {
					log.Printf("Non-retriable API error: %v", apiErr)
				}
				break
			}
			if c.logLevel <= logging.LogLevelWarning {
				log.Printf("Retriable API error on attempt %d: %v", attempt+1, apiErr)
			}
			continue
		}

		// For non-APIError errors (like network errors), retry
		if c.logLevel <= logging.LogLevelWarning {
			log.Printf("Network error on attempt %d: %v", attempt+1, err)
		}
		continue
	}

	if c.logLevel <= logging.LogLevelError {
		log.Printf("Failed to fetch user location after %d attempts: %v", c.maxRetries+1, lastErr)
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// doGetUserLocation performs a single attempt to fetch the user location
func (c *Client) doGetUserLocation(ctx context.Context) (*UserLocation, error) {
	req, err := http.NewRequest(http.MethodGet, c.url, nil)
	if err != nil {
		if c.logLevel <= logging.LogLevelError {
			log.Printf("Failed to create HTTP request: %v", err)
		}
		return nil, &Error{
			Retriable: false,
			Err:       fmt.Errorf("failed to create request: %w", err),
		}
	}

	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", fmt.Sprintf("mullvad-compass/%s", c.version))

	if c.logLevel <= logging.LogLevelDebug {
		log.Printf("Sending GET request to %s", c.url)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if c.logLevel <= logging.LogLevelError {
			log.Printf("HTTP request failed: %v", err)
		}
		return nil, fmt.Errorf("failed to fetch user location: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if c.logLevel <= logging.LogLevelDebug {
		log.Printf("Received HTTP %d response", resp.StatusCode)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		retriable := isRetriableStatusCode(resp.StatusCode)
		if c.logLevel <= logging.LogLevelWarning {
			log.Printf("Unexpected HTTP status code: %d (retriable: %v)", resp.StatusCode, retriable)
		}
		return nil, &Error{
			StatusCode: resp.StatusCode,
			Retriable:  retriable,
			Err:        fmt.Errorf("unexpected status code %d", resp.StatusCode),
		}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		if c.logLevel <= logging.LogLevelError {
			log.Printf("Unexpected content type: %s (expected application/json)", contentType)
		}
		return nil, &Error{
			Retriable: false,
			Err:       fmt.Errorf("unexpected content-type: %s (expected application/json)", contentType),
		}
	}

	var location UserLocation
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		if c.logLevel <= logging.LogLevelError {
			log.Printf("Failed to parse JSON response: %v", err)
		}
		return nil, &Error{
			Retriable: false,
			Err:       fmt.Errorf("failed to parse API response: %w", err),
		}
	}

	return &location, nil
}

// GetUserLocation is a convenience function that uses the default client
func GetUserLocation(ctx context.Context) (*UserLocation, error) {
	client := NewClient()
	return client.GetUserLocation(ctx)
}
