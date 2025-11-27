package ping

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

func TestPingLocationsWithFactory_EmptyLocations(t *testing.T) {
	factory := NewMockPingerFactory()

	result, err := LocationsWithFactory(context.Background(),
		[]relays.Location{},
		500,
		25,
		relays.IPv4,
		factory, logging.LogLevelError,
	)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d locations", len(result))
	}

	// Verify factory was NOT called (no need to create privileged socket for empty locations)
	calls := factory.GetCreatePingerCalls()
	if len(calls) != 0 {
		t.Errorf("Expected 0 CreatePinger calls for empty locations, got %d", len(calls))
	}
}

func TestPingLocationsWithFactory_SingleLocation(t *testing.T) {
	factory := NewMockPingerFactory()

	locations := []relays.Location{
		{
			IPv4Address: "1.2.3.4",
			Hostname:    "test-server",
			Country:     "Test",
			City:        "City",
		},
	}

	result, err := LocationsWithFactory(context.Background(),
		locations,
		500,
		25,
		relays.IPv4,
		factory, logging.LogLevelError,
	)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if result[0].Hostname != "test-server" {
		t.Errorf("Expected hostname 'test-server', got %s", result[0].Hostname)
	}

	if result[0].Latency == nil {
		t.Error("Expected latency to be set")
	} else if *result[0].Latency != 10.0 {
		t.Errorf("Expected latency 10.0ms, got %fms", *result[0].Latency)
	}

	// Verify the pinger was called
	pingers := factory.GetCreatedPingers()
	if len(pingers) != 1 {
		t.Fatalf("Expected 1 pinger created, got %d", len(pingers))
	}

	if pingers[0].GetPingCallCount() != 1 {
		t.Errorf("Expected 1 ping call, got %d", pingers[0].GetPingCallCount())
	}

	pingCalls := pingers[0].GetPingCalls()
	if len(pingCalls) != 1 {
		t.Fatalf("Expected 1 ping call recorded, got %d", len(pingCalls))
	}

	if pingCalls[0].IPAddr != "1.2.3.4" {
		t.Errorf("Expected IP 1.2.3.4, got %s", pingCalls[0].IPAddr)
	}

	if pingCalls[0].Timeout != 500*time.Millisecond {
		t.Errorf("Expected timeout 500ms, got %v", pingCalls[0].Timeout)
	}
}

func TestPingLocationsWithFactory_MultipleLocations(t *testing.T) {
	factory := NewMockPingerFactory()

	locations := []relays.Location{
		{IPv4Address: "1.1.1.1", Hostname: "server1"},
		{IPv4Address: "2.2.2.2", Hostname: "server2"},
		{IPv4Address: "3.3.3.3", Hostname: "server3"},
	}

	result, err := LocationsWithFactory(context.Background(),
		locations,
		500,
		25,
		relays.IPv4,
		factory, logging.LogLevelError,
	)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(result))
	}

	// Verify all locations have latency
	for _, loc := range result {
		if loc.Latency == nil {
			t.Errorf("Location %s has nil latency", loc.Hostname)
		}
	}

	// Verify pinger was called 3 times
	pingers := factory.GetCreatedPingers()
	if len(pingers) != 1 {
		t.Fatalf("Expected 1 pinger created, got %d", len(pingers))
	}

	if pingers[0].GetPingCallCount() != 3 {
		t.Errorf("Expected 3 ping calls, got %d", pingers[0].GetPingCallCount())
	}
}

func TestPingLocationsWithFactory_IPv6(t *testing.T) {
	factory := NewMockPingerFactory()

	locations := []relays.Location{
		{
			IPv6Address: "2001:db8::1",
			Hostname:    "ipv6-server",
		},
	}

	result, err := LocationsWithFactory(context.Background(),
		locations,
		500,
		25,
		relays.IPv6,
		factory, logging.LogLevelError,
	)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// Verify IPv6 address was used
	pingers := factory.GetCreatedPingers()
	if len(pingers) != 1 {
		t.Fatalf("Expected 1 pinger created, got %d", len(pingers))
	}

	pingCalls := pingers[0].GetPingCalls()
	if len(pingCalls) != 1 {
		t.Fatalf("Expected 1 ping call, got %d", len(pingCalls))
	}

	if pingCalls[0].IPAddr != "2001:db8::1" {
		t.Errorf("Expected IPv6 address 2001:db8::1, got %s", pingCalls[0].IPAddr)
	}

	// Verify factory was called with IPv6
	factoryCalls := factory.GetCreatePingerCalls()
	if len(factoryCalls) != 1 {
		t.Fatalf("Expected 1 factory call, got %d", len(factoryCalls))
	}

	if !factoryCalls[0].IPVersion.IsIPv6() {
		t.Error("Expected IPv6 version in factory call")
	}
}

func TestPingLocationsWithFactory_TimeoutSimulation(t *testing.T) {
	factory := NewMockPingerFactory()

	// Configure mock to return nil (timeout) for specific IPs
	mockPinger := NewMockPinger()
	mockPinger.PingFunc = func(_ context.Context, ipAddr string, _ time.Duration) *float64 {
		if ipAddr == "1.1.1.1" {
			latency := 50.0
			return &latency
		}
		// Simulate timeout
		return nil
	}
	factory.CreatePingerFunc = func(_ relays.IPVersion) (Pinger, error) {
		return mockPinger, nil
	}

	locations := []relays.Location{
		{IPv4Address: "1.1.1.1", Hostname: "reachable"},
		{IPv4Address: "2.2.2.2", Hostname: "unreachable"},
	}

	result, err := LocationsWithFactory(context.Background(),
		locations,
		500,
		25,
		relays.IPv4,
		factory, logging.LogLevelError,
	)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result))
	}

	// Find each result
	var reachable, unreachable *relays.Location
	for i := range result {
		switch result[i].Hostname {
		case "reachable":
			reachable = &result[i]
		case "unreachable":
			unreachable = &result[i]
		}
	}

	if reachable == nil || unreachable == nil {
		t.Fatal("Could not find both locations in result")
	}

	if reachable.Latency == nil {
		t.Error("Reachable server should have latency")
	} else if *reachable.Latency != 50.0 {
		t.Errorf("Expected latency 50.0ms, got %fms", *reachable.Latency)
	}

	if unreachable.Latency != nil {
		t.Errorf("Unreachable server should have nil latency, got %fms", *unreachable.Latency)
	}
}

func TestPingLocationsWithFactory_FactoryError(t *testing.T) {
	factory := NewMockPingerFactory()
	factory.CreatePingerErrFunc = func() error {
		return fmt.Errorf("simulated factory error")
	}

	locations := []relays.Location{
		{IPv4Address: "1.1.1.1", Hostname: "server1"},
	}

	_, err := LocationsWithFactory(context.Background(),
		locations,
		500,
		25,
		relays.IPv4,
		factory, logging.LogLevelError,
	)

	if err == nil {
		t.Fatal("Expected error from factory, got nil")
	}

	if err.Error() != "simulated factory error" {
		t.Errorf("Expected 'simulated factory error', got: %v", err)
	}
}

func TestPingLocationsWithFactory_WorkerPoolSize(t *testing.T) {
	factory := NewMockPingerFactory()

	// Create 100 locations
	locations := make([]relays.Location, 100)
	for i := range locations {
		locations[i] = relays.Location{
			IPv4Address: fmt.Sprintf("1.2.3.%d", i),
			Hostname:    fmt.Sprintf("server%d", i),
		}
	}

	result, err := LocationsWithFactory(context.Background(),
		locations,
		500,
		10, // Only 10 workers
		relays.IPv4,
		factory, logging.LogLevelError,
	)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 100 {
		t.Errorf("Expected 100 results, got %d", len(result))
	}

	// Verify all 100 locations were pinged
	pingers := factory.GetCreatedPingers()
	if len(pingers) != 1 {
		t.Fatalf("Expected 1 pinger created, got %d", len(pingers))
	}

	if pingers[0].GetPingCallCount() != 100 {
		t.Errorf("Expected 100 ping calls, got %d", pingers[0].GetPingCallCount())
	}
}

func TestPingLocationsWithFactory_PingerClosed(t *testing.T) {
	factory := NewMockPingerFactory()

	locations := []relays.Location{
		{IPv4Address: "1.1.1.1", Hostname: "server1"},
	}

	_, err := LocationsWithFactory(context.Background(),
		locations,
		500,
		25,
		relays.IPv4,
		factory, logging.LogLevelError,
	)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify pinger was closed
	pingers := factory.GetCreatedPingers()
	if len(pingers) != 1 {
		t.Fatalf("Expected 1 pinger created, got %d", len(pingers))
	}

	if !pingers[0].IsClosed() {
		t.Error("Expected pinger to be closed")
	}
}
