package protocol

import (
	"encoding/binary"
	"net"
)

// SubscribePayload represents a listener subscription request
type SubscribePayload struct {
	ListenerIPv6 [16]byte
	ListenerPort uint16
	Callsign     [16]byte
	Group        [32]byte // Multicast group name (e.g., "emergency", "community")
	SSMSource    [16]byte // SSM source IPv6 (all zeros = regular multicast)
}

// HeartbeatPayload represents a keepalive from listener
type HeartbeatPayload struct {
	ListenerIPv6 [16]byte
	Timestamp    uint64
}

// MarshalSubscribe encodes subscription payload to bytes
func MarshalSubscribe(sp *SubscribePayload) []byte {
	buf := make([]byte, 82) // 16 + 2 + 16 + 32 + 16

	copy(buf[0:16], sp.ListenerIPv6[:])
	binary.BigEndian.PutUint16(buf[16:18], sp.ListenerPort)
	copy(buf[18:34], sp.Callsign[:])
	copy(buf[34:66], sp.Group[:])
	copy(buf[66:82], sp.SSMSource[:])

	return buf
}

// UnmarshalSubscribe decodes subscription payload from bytes
func UnmarshalSubscribe(data []byte) (*SubscribePayload, error) {
	// Support both old (34 bytes) and new (82 bytes) formats
	if len(data) < 34 {
		return nil, ErrInvalidPayload
	}

	sp := &SubscribePayload{
		ListenerPort: binary.BigEndian.Uint16(data[16:18]),
	}

	copy(sp.ListenerIPv6[:], data[0:16])
	copy(sp.Callsign[:], data[18:34])

	// If new format with Group and SSMSource
	if len(data) >= 82 {
		copy(sp.Group[:], data[34:66])
		copy(sp.SSMSource[:], data[66:82])
	}

	return sp, nil
}

// MarshalHeartbeat encodes heartbeat payload to bytes
func MarshalHeartbeat(hp *HeartbeatPayload) []byte {
	buf := make([]byte, 24) // 16 + 8

	copy(buf[0:16], hp.ListenerIPv6[:])
	binary.BigEndian.PutUint64(buf[16:24], hp.Timestamp)

	return buf
}

// UnmarshalHeartbeat decodes heartbeat payload from bytes
func UnmarshalHeartbeat(data []byte) (*HeartbeatPayload, error) {
	if len(data) < 24 {
		return nil, ErrInvalidPayload
	}

	hp := &HeartbeatPayload{
		Timestamp: binary.BigEndian.Uint64(data[16:24]),
	}

	copy(hp.ListenerIPv6[:], data[0:16])

	return hp, nil
}

// Helper to convert net.IP to [16]byte
func IPv6ToBytes(ip net.IP) [16]byte {
	var result [16]byte
	copy(result[:], ip.To16())
	return result
}

// Helper to convert [16]byte to net.IP
func BytesToIPv6(b [16]byte) net.IP {
	return net.IP(b[:])
}

// Helper to get callsign as string
func GetCallsignString(callsign [16]byte) string {
	length := 0
	for i, b := range callsign {
		if b == 0 {
			length = i
			break
		}
		if i == len(callsign)-1 {
			length = len(callsign)
		}
	}
	return string(callsign[:length])
}

// Helper to get group name as string
func GetGroupString(group [32]byte) string {
	length := 0
	for i, b := range group {
		if b == 0 {
			length = i
			break
		}
		if i == len(group)-1 {
			length = len(group)
		}
	}
	return string(group[:length])
}

// Helper to convert string to group bytes
func StringToGroup(s string) [32]byte {
	var result [32]byte
	copy(result[:], []byte(s))
	return result
}

// IsZeroIPv6 checks if an IPv6 address is all zeros
func IsZeroIPv6(ip [16]byte) bool {
	for _, b := range ip {
		if b != 0 {
			return false
		}
	}
	return true
}
