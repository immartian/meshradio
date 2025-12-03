package protocol

import (
	"encoding/binary"
	"errors"
	"time"
)

// Protocol version
const Version uint8 = 0x01

// Packet types
const (
	PacketTypeBeacon         uint8 = 0x00
	PacketTypeAudio          uint8 = 0x01
	PacketTypeMetadata       uint8 = 0x02
	PacketTypeCallCQ         uint8 = 0x03
	PacketTypeCallSelective  uint8 = 0x04
	PacketTypeDiscoveryReq   uint8 = 0x05
	PacketTypeDiscoveryResp  uint8 = 0x06
	PacketTypeSignalReport   uint8 = 0x09
	PacketTypeEmergency      uint8 = 0x0A

	// Subscription-based streaming (MVP)
	PacketTypeSubscribe      uint8 = 0x10
	PacketTypeHeartbeat      uint8 = 0x11
	PacketTypeUnsubscribe    uint8 = 0x12
)

// Packet flags
const (
	FlagEncrypted uint8 = 0x01
	FlagPriority  uint8 = 0x02
	FlagCompressed uint8 = 0x04
)

// Header size in bytes
const HeaderSize = 64

// Packet represents a MeshRadio protocol packet
type Packet struct {
	Version        uint8
	Type           uint8
	Flags          uint8
	PayloadLength  uint16
	Timestamp      int64
	SourceIPv6     [16]byte
	Callsign       [16]byte
	SequenceNum    uint8
	SignalQuality  uint8
	Reserved       uint8
	Payload        []byte
}

// NewPacket creates a new packet with given type and payload
func NewPacket(packetType uint8, sourceIPv6 [16]byte, callsign string, payload []byte) *Packet {
	var callsignBytes [16]byte
	copy(callsignBytes[:], []byte(callsign))

	return &Packet{
		Version:       Version,
		Type:          packetType,
		Flags:         0,
		PayloadLength: uint16(len(payload)),
		Timestamp:     time.Now().UnixMilli(),
		SourceIPv6:    sourceIPv6,
		Callsign:      callsignBytes,
		SequenceNum:   0,
		SignalQuality: 0,
		Reserved:      0,
		Payload:       payload,
	}
}

// Marshal encodes the packet to bytes
func (p *Packet) Marshal() ([]byte, error) {
	totalSize := HeaderSize + len(p.Payload)
	buf := make([]byte, totalSize)

	// Version and Type (4 bits each)
	buf[0] = (p.Version << 4) | (p.Type & 0x0F)

	// Flags
	buf[1] = p.Flags

	// Payload Length
	binary.BigEndian.PutUint16(buf[2:4], p.PayloadLength)

	// Timestamp
	binary.BigEndian.PutUint64(buf[4:12], uint64(p.Timestamp))

	// Source IPv6 (16 bytes)
	copy(buf[12:28], p.SourceIPv6[:])

	// Callsign (16 bytes)
	copy(buf[28:44], p.Callsign[:])

	// Sequence Number
	buf[44] = p.SequenceNum

	// Signal Quality
	buf[45] = p.SignalQuality

	// Reserved
	buf[46] = p.Reserved

	// Payload
	copy(buf[HeaderSize:], p.Payload)

	return buf, nil
}

// Unmarshal decodes bytes into a packet
func Unmarshal(data []byte) (*Packet, error) {
	if len(data) < HeaderSize {
		return nil, errors.New("packet too small")
	}

	p := &Packet{}

	// Version and Type
	p.Version = (data[0] >> 4) & 0x0F
	p.Type = data[0] & 0x0F

	// Flags
	p.Flags = data[1]

	// Payload Length
	p.PayloadLength = binary.BigEndian.Uint16(data[2:4])

	// Timestamp
	p.Timestamp = int64(binary.BigEndian.Uint64(data[4:12]))

	// Source IPv6
	copy(p.SourceIPv6[:], data[12:28])

	// Callsign
	copy(p.Callsign[:], data[28:44])

	// Sequence Number
	p.SequenceNum = data[44]

	// Signal Quality
	p.SignalQuality = data[45]

	// Reserved
	p.Reserved = data[46]

	// Payload
	if len(data) > HeaderSize {
		p.Payload = make([]byte, len(data)-HeaderSize)
		copy(p.Payload, data[HeaderSize:])
	}

	// Validate payload length
	if len(p.Payload) != int(p.PayloadLength) {
		return nil, errors.New("payload length mismatch")
	}

	return p, nil
}

// GetCallsign returns the callsign as a string
func (p *Packet) GetCallsign() string {
	// Find null terminator or use full length
	length := 0
	for i, b := range p.Callsign {
		if b == 0 {
			length = i
			break
		}
		if i == len(p.Callsign)-1 {
			length = len(p.Callsign)
		}
	}
	return string(p.Callsign[:length])
}
