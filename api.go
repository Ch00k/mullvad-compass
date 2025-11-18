package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const mullvadAPIURL = "https://am.i.mullvad.net/json"

// UserLocation represents the response from Mullvad's location API
type UserLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
}

// GetUserLocation fetches the user's current geographic location from Mullvad API
func GetUserLocation() (*UserLocation, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(mullvadAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user location: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var location UserLocation
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &location, nil
}
