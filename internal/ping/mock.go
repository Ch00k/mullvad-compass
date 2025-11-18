package ping

import (
	"context"
	"sync"
	"time"

	"github.com/Ch00k/mullvad-compass/internal/relays"
)

// MockPinger is a mock implementation of Pinger for testing
type MockPinger struct {
	mu            sync.Mutex
	PingFunc      func(ctx context.Context, ipAddr string, timeout time.Duration) *float64
	CloseFunc     func() error
	PingCallCount int
	PingCalls     []Call
	Closed        bool
}

// Call records a call to Ping
type Call struct {
	IPAddr  string
	Timeout time.Duration
}

// NewMockPinger creates a new mock pinger with default behavior
func NewMockPinger() *MockPinger {
	return &MockPinger{
		PingFunc: func(_ context.Context, _ string, _ time.Duration) *float64 {
			// Default: return successful ping with 10ms latency
			latency := 10.0
			return &latency
		},
		CloseFunc: func() error {
			return nil
		},
		PingCalls: make([]Call, 0),
	}
}

// Ping implements the Pinger interface
func (m *MockPinger) Ping(ctx context.Context, ipAddr string, timeout time.Duration) *float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PingCallCount++
	m.PingCalls = append(m.PingCalls, Call{
		IPAddr:  ipAddr,
		Timeout: timeout,
	})

	if m.PingFunc != nil {
		return m.PingFunc(ctx, ipAddr, timeout)
	}
	return nil
}

// Close implements the Pinger interface
func (m *MockPinger) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Closed = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// GetPingCallCount returns the number of times Ping was called
func (m *MockPinger) GetPingCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.PingCallCount
}

// GetPingCalls returns all recorded Ping calls
func (m *MockPinger) GetPingCalls() []Call {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Call(nil), m.PingCalls...)
}

// IsClosed returns whether Close was called
func (m *MockPinger) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Closed
}

// MockPingerFactory is a mock implementation of PingerFactory for testing
type MockPingerFactory struct {
	mu                  sync.Mutex
	CreatePingerFunc    func(ipVersion relays.IPVersion) (Pinger, error)
	CreatePingerCalls   []CreatePingerCall
	CreatedPingers      []*MockPinger
	CreatePingerErrFunc func() error
}

// CreatePingerCall records a call to CreatePinger
type CreatePingerCall struct {
	IPVersion relays.IPVersion
}

// NewMockPingerFactory creates a new mock pinger factory
func NewMockPingerFactory() *MockPingerFactory {
	return &MockPingerFactory{
		CreatePingerCalls: make([]CreatePingerCall, 0),
		CreatedPingers:    make([]*MockPinger, 0),
	}
}

// CreatePinger implements the PingerFactory interface
func (f *MockPingerFactory) CreatePinger(ipVersion relays.IPVersion) (Pinger, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.CreatePingerCalls = append(f.CreatePingerCalls, CreatePingerCall{
		IPVersion: ipVersion,
	})

	if f.CreatePingerErrFunc != nil {
		if err := f.CreatePingerErrFunc(); err != nil {
			return nil, err
		}
	}

	if f.CreatePingerFunc != nil {
		return f.CreatePingerFunc(ipVersion)
	}

	// Default: create a new mock pinger
	mockPinger := NewMockPinger()
	f.CreatedPingers = append(f.CreatedPingers, mockPinger)
	return mockPinger, nil
}

// GetCreatePingerCalls returns all recorded CreatePinger calls
func (f *MockPingerFactory) GetCreatePingerCalls() []CreatePingerCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]CreatePingerCall(nil), f.CreatePingerCalls...)
}

// GetCreatedPingers returns all created mock pingers
func (f *MockPingerFactory) GetCreatedPingers() []*MockPinger {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*MockPinger(nil), f.CreatedPingers...)
}
