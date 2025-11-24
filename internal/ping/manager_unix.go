//go:build !windows

package ping

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/icmp"
	"github.com/Ch00k/mullvad-compass/internal/relays"
	xicmp "golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// pingResponse contains the response from a ping operation
type pingResponse struct {
	peerIP net.IP
	rtt    time.Duration
}

// socketManager manages shared ICMP sockets for IPv4 and IPv6
type socketManager struct {
	conn       *xicmp.PacketConn
	network    string
	protocol   int
	seqCounter atomic.Int32
	inFlight   sync.Map // map[int]chan *pingResponse
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// newSocketManager creates a new socket manager for the given IP version
func newSocketManager(ipVersion relays.IPVersion) (*socketManager, error) {
	conn, network, err := icmp.Listen(ipVersion)
	if err != nil {
		return nil, err
	}

	var protocol int
	if ipVersion.IsIPv6() {
		protocol = protocolICMPv6
	} else {
		protocol = protocolICMP
	}

	ctx, cancel := context.WithCancel(context.Background())
	mgr := &socketManager{
		conn:     conn,
		network:  network,
		protocol: protocol,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start reader goroutine
	mgr.wg.Add(1)
	go mgr.reader()

	return mgr, nil
}

// allocateSeq allocates a unique sequence number
func (m *socketManager) allocateSeq() int {
	return int(m.seqCounter.Add(1))
}

// reader continuously reads ICMP responses and routes them to waiting goroutines
func (m *socketManager) reader() {
	defer m.wg.Done()

	buffer := make([]byte, 1500)
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		// Set a read deadline to periodically check context
		_ = m.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		n, peer, err := m.conn.ReadFrom(buffer)
		if err != nil {
			// Check if it's a timeout error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			// Context cancelled or other error
			if m.ctx.Err() != nil {
				return
			}
			continue
		}

		// Parse ICMP message
		msg, err := xicmp.ParseMessage(m.protocol, buffer[:n])
		if err != nil {
			continue
		}

		// Verify it's an echo reply
		var isEchoReply bool
		if m.protocol == protocolICMPv6 {
			isEchoReply = msg.Type == ipv6.ICMPTypeEchoReply
		} else {
			isEchoReply = msg.Type == ipv4.ICMPTypeEchoReply
		}
		if !isEchoReply {
			continue
		}

		// Extract sequence number
		echo, ok := msg.Body.(*xicmp.Echo)
		if !ok {
			continue
		}

		// Extract peer IP
		var peerIP net.IP
		switch addr := peer.(type) {
		case *net.UDPAddr:
			peerIP = addr.IP
		case *net.IPAddr:
			peerIP = addr.IP
		default:
			continue
		}

		// Route to waiting goroutine
		if chInterface, ok := m.inFlight.LoadAndDelete(echo.Seq); ok {
			ch := chInterface.(chan *pingResponse)
			select {
			case ch <- &pingResponse{peerIP: peerIP, rtt: 0}: // RTT will be calculated by sender
			default:
				// Channel full or closed, ignore
			}
		}
	}
}

// Ping sends an ICMP echo request and waits for a response
func (m *socketManager) Ping(ctx context.Context, ipAddr string, timeout time.Duration) *float64 {
	// Parse IP address
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return nil
	}

	// Allocate sequence number
	seq := m.allocateSeq()

	// Create response channel
	respChan := make(chan *pingResponse, 1)
	m.inFlight.Store(seq, respChan)

	// Ensure cleanup
	defer func() {
		m.inFlight.Delete(seq)
		close(respChan)
	}()

	// Build ICMP message
	var msg xicmp.Message
	if m.protocol == protocolICMPv6 {
		msg = xicmp.Message{
			Type: ipv6.ICMPTypeEchoRequest,
			Code: 0,
			Body: &xicmp.Echo{
				ID:   1,
				Seq:  seq,
				Data: []byte("mullvad-compass"),
			},
		}
	} else {
		msg = xicmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &xicmp.Echo{
				ID:   1,
				Seq:  seq,
				Data: []byte("mullvad-compass"),
			},
		}
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return nil
	}

	// Resolve destination (UDP ICMP uses UDP addresses)
	var dst net.Addr
	if m.protocol == protocolICMPv6 {
		addr, err := net.ResolveUDPAddr("udp6", "["+ipAddr+"]:0")
		if err != nil {
			return nil
		}
		dst = addr
	} else {
		addr, err := net.ResolveUDPAddr("udp4", ipAddr+":0")
		if err != nil {
			return nil
		}
		dst = addr
	}

	// Send ping
	start := time.Now()
	_, err = m.conn.WriteTo(msgBytes, dst)
	if err != nil {
		return nil
	}

	// Wait for response with timeout or context cancellation
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case resp := <-respChan:
		if resp == nil {
			return nil
		}
		// Verify response came from the correct IP
		if !resp.peerIP.Equal(ip) {
			return nil
		}
		latencyMs := float64(time.Since(start).Microseconds()) / 1000.0
		return &latencyMs
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return nil
	}
}

// Close shuts down the socket manager and waits for reader goroutine to exit
func (m *socketManager) Close() error {
	m.cancel()
	// Close the connection first to unblock the reader immediately
	err := m.conn.Close()
	m.wg.Wait()
	return err
}
