package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_GetUserLocation_Success(t *testing.T) {
	expectedLocation := UserLocation{
		IP:            "1.2.3.4",
		Latitude:      40.7128,
		Longitude:     -74.0060,
		Country:       "USA",
		City:          "New York",
		MullvadExitIP: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent is set
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent header not set")
		}
		if !strings.Contains(r.Header.Get("User-Agent"), "mullvad-compass") {
			t.Errorf("User-Agent should contain 'mullvad-compass', got: %s", r.Header.Get("User-Agent"))
		}
		// Verify default version is used
		if r.Header.Get("User-Agent") != "mullvad-compass/dev" {
			t.Errorf("Expected User-Agent 'mullvad-compass/dev', got: %s", r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedLocation)
	}))
	defer server.Close()

	client := NewClient(WithURL(server.URL))
	location, err := client.GetUserLocation(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if location.IP != expectedLocation.IP {
		t.Errorf("Expected IP %s, got %s", expectedLocation.IP, location.IP)
	}
	if location.Latitude != expectedLocation.Latitude {
		t.Errorf("Expected Latitude %f, got %f", expectedLocation.Latitude, location.Latitude)
	}
	if location.Longitude != expectedLocation.Longitude {
		t.Errorf("Expected Longitude %f, got %f", expectedLocation.Longitude, location.Longitude)
	}
}

func TestClient_GetUserLocation_WrongContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>Not JSON</html>"))
	}))
	defer server.Close()

	client := NewClient(WithURL(server.URL))
	_, err := client.GetUserLocation(context.Background())

	if err == nil {
		t.Fatal("Expected error for wrong content type, got nil")
	}
	if !strings.Contains(err.Error(), "content-type") && !strings.Contains(err.Error(), "Content-Type") {
		t.Errorf("Expected error to mention content-type, got: %v", err)
	}
}

func TestClient_GetUserLocation_NonOKStatus(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"Bad Request", http.StatusBadRequest},
		{"Internal Server Error", http.StatusInternalServerError},
		{"Service Unavailable", http.StatusServiceUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			client := NewClient(WithURL(server.URL), WithMaxRetries(0))
			_, err := client.GetUserLocation(context.Background())

			if err == nil {
				t.Fatalf("Expected error for status %d, got nil", tc.statusCode)
			}
		})
	}
}

func TestClient_GetUserLocation_RetryOnTransientFailure(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(UserLocation{
			IP:       "1.2.3.4",
			Latitude: 40.7128,
		})
	}))
	defer server.Close()

	client := NewClient(
		WithURL(server.URL),
		WithMaxRetries(3),
		WithRetryDelay(1*time.Millisecond),
	)
	location, err := client.GetUserLocation(context.Background())
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}
	if location.IP != "1.2.3.4" {
		t.Errorf("Expected IP 1.2.3.4, got %s", location.IP)
	}
	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestClient_GetUserLocation_ExhaustedRetries(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewClient(
		WithURL(server.URL),
		WithMaxRetries(2),
		WithRetryDelay(1*time.Millisecond),
	)
	_, err := client.GetUserLocation(context.Background())

	if err == nil {
		t.Fatal("Expected error after exhausting retries, got nil")
	}
	if attemptCount != 3 { // Initial attempt + 2 retries
		t.Errorf("Expected 3 attempts (1 initial + 2 retries), got %d", attemptCount)
	}
}

func TestClient_GetUserLocation_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(WithURL(server.URL))
	_, err := client.GetUserLocation(context.Background())

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestClient_GetUserLocation_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithURL(server.URL),
		WithTimeout(10*time.Millisecond),
		WithMaxRetries(0),
	)
	_, err := client.GetUserLocation(context.Background())

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestClient_GetUserLocation_CustomVersion(t *testing.T) {
	customVersion := "1.2.3"
	expectedUserAgent := "mullvad-compass/1.2.3"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		if userAgent != expectedUserAgent {
			t.Errorf("Expected User-Agent '%s', got: %s", expectedUserAgent, userAgent)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(UserLocation{
			IP:       "1.2.3.4",
			Latitude: 40.7128,
		})
	}))
	defer server.Close()

	client := NewClient(
		WithURL(server.URL),
		WithVersion(customVersion),
	)
	_, err := client.GetUserLocation(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}
