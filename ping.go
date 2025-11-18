package main

import (
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP = 1
)

// PingResult contains the result of a ping operation
type PingResult struct {
	Location *Location
	Latency  *float64
}

// PingLocations pings all locations concurrently and updates their latency values
func PingLocations(locations []Location, timeout, workers int) ([]Location, error) {
	workChan := make(chan *Location, len(locations))
	resultChan := make(chan PingResult, len(locations))

	to := time.Duration(timeout) * time.Millisecond

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pingWorker(workChan, resultChan, to)
		}()
	}

	// Send locations to workers in a separate goroutine
	go func() {
		for i := range locations {
			workChan <- &locations[i]
		}
		close(workChan)
	}()

	// Wait for all workers to finish, then close results channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]Location, 0, len(locations))
	for result := range resultChan {
		result.Location.Latency = result.Latency
		results = append(results, *result.Location)
	}

	return results, nil
}

// pingWorker processes locations from the work channel
func pingWorker(workChan <-chan *Location, resultChan chan<- PingResult, timeout time.Duration) {
	for loc := range workChan {
		latency := ping(loc.IPv4Address, timeout)
		resultChan <- PingResult{
			Location: loc,
			Latency:  latency,
		}
	}
}

// listenICMP tries to create an ICMP connection, attempting raw ICMP first, then UDP fallback
func listenICMP() (*icmp.PacketConn, string, error) {
	// Try raw ICMP first (requires privileges but gives more accurate results)
	if c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0"); err == nil {
		return c, "ip4:icmp", nil
	}

	// Fallback to UDP datagram (unprivileged on macOS and Linux with net.ipv4.ping_group_range)
	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return nil, "", err
	}
	return c, "udp4", nil
}

// ping sends an ICMP echo request with its own dedicated socket
func ping(ipAddr string, timeout time.Duration) *float64 {
	// Create dedicated ICMP connection for this ping
	conn, network, err := listenICMP()
	if err != nil {
		return nil
	}
	defer func() { _ = conn.Close() }()

	// Create ICMP echo request
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1, // ID doesn't matter for SOCK_DGRAM, kernel manages it
			Seq:  1,
			Data: []byte("mullvad-compass"),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return nil
	}

	// Resolve destination
	var dst net.Addr
	if network == "udp4" {
		addr, err := net.ResolveUDPAddr("udp4", ipAddr+":0")
		if err != nil {
			return nil
		}
		dst = addr
	} else {
		addr, err := net.ResolveIPAddr("ip4", ipAddr)
		if err != nil {
			return nil
		}
		dst = addr
	}

	// Send ping
	start := time.Now()
	_, err = conn.WriteTo(msgBytes, dst)
	if err != nil {
		return nil
	}

	// Set read deadline for timeout
	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	// Wait for reply - may need to read multiple packets if intermediate routers respond
	reply := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(reply)
		if err != nil {
			return nil // Timeout or error
		}

		// Parse reply to verify it's valid
		parsedMsg, err := icmp.ParseMessage(protocolICMP, reply[:n])
		if err != nil {
			continue // Invalid message, keep reading
		}

		// Verify it's an echo reply
		if parsedMsg.Type != ipv4.ICMPTypeEchoReply {
			continue // Not an echo reply, keep reading
		}

		// Verify the reply came from the IP we pinged
		var peerIP string
		switch addr := peer.(type) {
		case *net.UDPAddr:
			peerIP = addr.IP.String()
		case *net.IPAddr:
			peerIP = addr.IP.String()
		default:
			continue // Unknown address type, keep reading
		}

		if peerIP != ipAddr {
			continue // Reply from wrong IP (intermediate router), keep reading
		}

		// Valid reply from correct IP
		latencyMs := float64(time.Since(start).Microseconds()) / 1000.0
		return &latencyMs
	}
}
