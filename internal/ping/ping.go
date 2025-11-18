// Package ping provides functions for pinging VPN servers using ICMP.
package ping

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/logging"
	"github.com/Ch00k/mullvad-compass/internal/relays"
)

const (
	protocolICMP   = 1
	protocolICMPv6 = 58
)

// Result contains the result of a ping operation
type Result struct {
	Location *relays.Location
	Latency  *float64
}

// Locations pings all locations concurrently and updates their latency values
func Locations(
	ctx context.Context,
	locations []relays.Location,
	timeout, workers int,
	ipVersion relays.IPVersion,
) ([]relays.Location, error) {
	return LocationsWithFactory(
		ctx,
		locations,
		timeout,
		workers,
		ipVersion,
		NewDefaultPingerFactory(),
		logging.LogLevelError,
	)
}

// LocationsWithFactory pings all locations using a provided pinger factory
func LocationsWithFactory(
	ctx context.Context,
	locations []relays.Location,
	timeout, workers int,
	ipVersion relays.IPVersion,
	factory PingerFactory,
	logLevel logging.LogLevel,
) ([]relays.Location, error) {
	if logLevel <= logging.LogLevelInfo {
		log.Printf(
			"Starting to ping %d locations with %d workers (timeout: %dms, IP version: %s)",
			len(locations),
			workers,
			timeout,
			ipVersion,
		)
	}

	// Create pinger for the specified IP version
	start := time.Now()
	pinger, err := factory.CreatePinger(ipVersion)
	if err != nil {
		if logLevel <= logging.LogLevelError {
			log.Printf("Failed to create pinger: %v", err)
		}
		return nil, err
	}
	defer func() { _ = pinger.Close() }()
	if logLevel <= logging.LogLevelDebug {
		log.Printf("Socket creation completed in %v", time.Since(start))
	}

	workChan := make(chan *relays.Location, len(locations))
	resultChan := make(chan Result, len(locations))

	to := time.Duration(timeout) * time.Millisecond

	// Start worker pool (don't spin up more workers than locations)
	numWorkers := workers
	if numWorkers > len(locations) {
		numWorkers = len(locations)
	}
	workerStart := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pingWorker(ctx, workChan, resultChan, to, pinger, ipVersion)
		}()
	}
	if logLevel <= logging.LogLevelDebug {
		log.Printf("Worker pool startup (%d workers) completed in %v", numWorkers, time.Since(workerStart))
	}

	// Send locations to workers in a separate goroutine
	go func() {
		for i := range locations {
			select {
			case <-ctx.Done():
				close(workChan)
				return
			case workChan <- &locations[i]:
			}
		}
		close(workChan)
	}()

	// Wait for all workers to finish, then close results channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	collectStart := time.Now()
	results := make([]relays.Location, 0, len(locations))
	var successCount, failCount int
	for result := range resultChan {
		result.Location.Latency = result.Latency
		results = append(results, *result.Location)
		if result.Latency != nil {
			successCount++
		} else {
			failCount++
		}
	}
	if logLevel <= logging.LogLevelDebug {
		log.Printf("Result collection completed in %v", time.Since(collectStart))
	}

	if logLevel <= logging.LogLevelInfo {
		log.Printf("Ping completed: %d successful, %d failed out of %d total", successCount, failCount, len(results))
	}

	// Check if context was cancelled
	if ctx.Err() != nil {
		if logLevel <= logging.LogLevelWarning {
			log.Printf("Ping operation cancelled: %v", ctx.Err())
		}
		return results, ctx.Err()
	}

	return results, nil
}

// pingWorker processes locations from the work channel
func pingWorker(
	ctx context.Context,
	workChan <-chan *relays.Location,
	resultChan chan<- Result,
	timeout time.Duration,
	pinger Pinger,
	ipVersion relays.IPVersion,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case loc, ok := <-workChan:
			if !ok {
				return
			}
			var ipAddr string
			if ipVersion.IsIPv6() {
				ipAddr = loc.IPv6Address
			} else {
				ipAddr = loc.IPv4Address
			}
			latency := pinger.Ping(ctx, ipAddr, timeout)
			select {
			case <-ctx.Done():
				return
			case resultChan <- Result{
				Location: loc,
				Latency:  latency,
			}:
			}
		}
	}
}
