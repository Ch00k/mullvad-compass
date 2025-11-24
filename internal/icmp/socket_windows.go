//go:build windows

package icmp

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	iphlpapi = windows.NewLazyDLL("iphlpapi.dll")

	procIcmpCreateFile  = iphlpapi.NewProc("IcmpCreateFile")
	procIcmpSendEcho    = iphlpapi.NewProc("IcmpSendEcho")
	procIcmpCloseHandle = iphlpapi.NewProc("IcmpCloseHandle")
	procIcmp6CreateFile = iphlpapi.NewProc("Icmp6CreateFile")
	procIcmp6SendEcho2  = iphlpapi.NewProc("Icmp6SendEcho2")
)

const (
	invalidHandleValue = ^uintptr(0)
)

// Windows IP status codes
const (
	IPSuccess             = 0
	IPBufTooSmall         = 11001
	IPDestNetUnreachable  = 11002
	IPDestHostUnreachable = 11003
	IPDestProtUnreachable = 11004
	IPDestPortUnreachable = 11005
	IPNoResources         = 11006
	IPBadOption           = 11007
	IPHWError             = 11008
	IPPacketTooBig        = 11009
	IPReqTimedOut         = 11010
	IPBadReq              = 11011
	IPBadRoute            = 11012
	IPTTLExpiredTransit   = 11013
	IPTTLExpiredReassem   = 11014
	IPParamProblem        = 11015
	IPSourceQuench        = 11016
	IPOptionTooBig        = 11017
	IPBadDestination      = 11018
)

// Handle represents a Windows ICMP handle
type Handle uintptr

// IPOptionInformation matches the Windows IP_OPTION_INFORMATION structure
type IPOptionInformation struct {
	TTL         uint8
	TOS         uint8
	Flags       uint8
	OptionsSize uint8
	OptionsData uintptr
}

// IcmpEchoReply matches the Windows ICMP_ECHO_REPLY structure
type IcmpEchoReply struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       IPOptionInformation
}

// Icmp6EchoReply matches the Windows ICMPV6_ECHO_REPLY structure
// Total size with alignment: 36 bytes (26 + 2 padding + 4 + 4)
type Icmp6EchoReply struct {
	Address       IPV6AddressEx // 26 bytes
	_padding      uint16        // 2 bytes padding for alignment
	Status        uint32        // 4 bytes
	RoundTripTime uint32        // 4 bytes
}

// IPV6AddressEx matches the Windows IPV6_ADDRESS_EX structure
// Total size: 26 bytes
type IPV6AddressEx struct {
	Port     uint16    // sin6_port (2 bytes)
	FlowInfo uint32    // sin6_flowinfo (4 bytes)
	Addr     [8]uint16 // sin6_addr[8] (16 bytes)
	ScopeID  uint32    // sin6_scope_id (4 bytes)
}

// SockAddrIn6 matches the Windows SOCKADDR_IN6 structure
type SockAddrIn6 struct {
	Family   uint16
	Port     uint16
	FlowInfo uint32
	Addr     [16]byte
	ScopeID  uint32
}

const (
	afInet6 = 23
)

// IcmpCreateFile creates an IPv4 ICMP handle
func IcmpCreateFile() (Handle, error) {
	ret, _, err := procIcmpCreateFile.Call()
	if ret == invalidHandleValue {
		if err != nil {
			return 0, fmt.Errorf("IcmpCreateFile failed: %w", err)
		}
		return 0, fmt.Errorf("IcmpCreateFile failed")
	}
	return Handle(ret), nil
}

// Icmp6CreateFile creates an IPv6 ICMP handle
func Icmp6CreateFile() (Handle, error) {
	ret, _, err := procIcmp6CreateFile.Call()
	if ret == invalidHandleValue {
		if err != nil {
			return 0, fmt.Errorf("Icmp6CreateFile failed: %w", err)
		}
		return 0, fmt.Errorf("Icmp6CreateFile failed")
	}
	return Handle(ret), nil
}

// IcmpCloseHandle closes an ICMP handle
func IcmpCloseHandle(handle Handle) error {
	ret, _, err := procIcmpCloseHandle.Call(uintptr(handle))
	if ret == 0 {
		if err != nil {
			return fmt.Errorf("IcmpCloseHandle failed: %w", err)
		}
		return fmt.Errorf("IcmpCloseHandle failed")
	}
	return nil
}

// IcmpSendEcho sends an IPv4 ICMP echo request and waits for reply
func IcmpSendEcho(handle Handle, destAddr uint32, requestData []byte, timeout time.Duration) (*IcmpEchoReply, error) {
	if len(requestData) == 0 {
		requestData = []byte("mullvad-compass")
	}

	timeoutMs := uint32(timeout.Milliseconds())

	// Allocate reply buffer
	// Size must be at least sizeof(ICMP_ECHO_REPLY) + requestData size + 8 bytes for padding
	replyBufSize := unsafe.Sizeof(IcmpEchoReply{}) + uintptr(len(requestData)) + 8
	replyBuf := make([]byte, replyBufSize)

	ret, _, err := procIcmpSendEcho.Call(
		uintptr(handle),
		uintptr(destAddr),
		uintptr(unsafe.Pointer(&requestData[0])),
		uintptr(len(requestData)),
		0, // IP options (NULL)
		uintptr(unsafe.Pointer(&replyBuf[0])),
		uintptr(len(replyBuf)),
		uintptr(timeoutMs),
	)

	if ret == 0 {
		// Parse reply even on failure to get status code
		reply := (*IcmpEchoReply)(unsafe.Pointer(&replyBuf[0]))
		if reply.Status != IPSuccess && reply.Status != 0 {
			return nil, fmt.Errorf("IcmpSendEcho failed with status: %s (syscall error: %w)", IPStatusToString(reply.Status), err)
		}
		return nil, fmt.Errorf("IcmpSendEcho failed: %w", err)
	}

	// Parse reply
	reply := (*IcmpEchoReply)(unsafe.Pointer(&replyBuf[0]))
	return reply, nil
}

// Icmp6SendEcho2 sends an IPv6 ICMP echo request and waits for reply
func Icmp6SendEcho2(handle Handle, destAddr net.IP, requestData []byte, timeout time.Duration) (*Icmp6EchoReply, error) {
	if len(requestData) == 0 {
		requestData = []byte("mullvad-compass")
	}

	timeoutMs := uint32(timeout.Milliseconds())

	// Source address - use unspecified address (::) to let system choose
	var srcSockAddr SockAddrIn6
	srcSockAddr.Family = afInet6
	// Leave Addr as zero (::) to let Windows select source address

	// Convert destination IP to SOCKADDR_IN6
	var destSockAddr SockAddrIn6
	destSockAddr.Family = afInet6
	copy(destSockAddr.Addr[:], destAddr.To16())

	// Allocate reply buffer
	replyBufSize := unsafe.Sizeof(Icmp6EchoReply{}) + uintptr(len(requestData)) + 8
	replyBuf := make([]byte, replyBufSize)

	ret, _, err := procIcmp6SendEcho2.Call(
		uintptr(handle),
		0, // Event (NULL for synchronous)
		0, // ApcRoutine (NULL)
		0, // ApcContext (NULL)
		uintptr(unsafe.Pointer(&srcSockAddr)), // Source address (required)
		uintptr(unsafe.Pointer(&destSockAddr)),
		uintptr(unsafe.Pointer(&requestData[0])),
		uintptr(len(requestData)),
		0, // IP options (NULL)
		uintptr(unsafe.Pointer(&replyBuf[0])),
		uintptr(len(replyBuf)),
		uintptr(timeoutMs),
	)

	if ret == 0 {
		// Parse reply even on failure to get status code
		reply := (*Icmp6EchoReply)(unsafe.Pointer(&replyBuf[0]))
		if reply.Status != IPSuccess && reply.Status != 0 {
			return nil, fmt.Errorf("Icmp6SendEcho2 failed with status: %s (syscall error: %w)", IPStatusToString(reply.Status), err)
		}
		return nil, fmt.Errorf("Icmp6SendEcho2 failed: %w", err)
	}

	// Parse reply
	reply := (*Icmp6EchoReply)(unsafe.Pointer(&replyBuf[0]))
	return reply, nil
}

// IPv4ToUint32 converts an IPv4 address to uint32 in little-endian byte order for Windows
func IPv4ToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	// Windows expects little-endian byte order
	return binary.LittleEndian.Uint32(ip)
}

// IPStatusToString converts an IP status code to a human-readable string
func IPStatusToString(status uint32) string {
	switch status {
	case IPSuccess:
		return "Success"
	case IPBufTooSmall:
		return "Reply buffer too small"
	case IPDestNetUnreachable:
		return "Destination network unreachable"
	case IPDestHostUnreachable:
		return "Destination host unreachable"
	case IPDestProtUnreachable:
		return "Destination protocol unreachable"
	case IPDestPortUnreachable:
		return "Destination port unreachable"
	case IPNoResources:
		return "Insufficient IP resources"
	case IPBadOption:
		return "Bad IP option"
	case IPHWError:
		return "Hardware error"
	case IPPacketTooBig:
		return "Packet too big"
	case IPReqTimedOut:
		return "Request timed out"
	case IPBadReq:
		return "Bad request"
	case IPBadRoute:
		return "Bad route"
	case IPTTLExpiredTransit:
		return "TTL expired in transit"
	case IPTTLExpiredReassem:
		return "TTL expired in reassembly"
	case IPParamProblem:
		return "Parameter problem"
	case IPSourceQuench:
		return "Source quench"
	case IPOptionTooBig:
		return "IP option too big"
	case IPBadDestination:
		return "Bad destination"
	default:
		return fmt.Sprintf("Unknown status: %d", status)
	}
}
