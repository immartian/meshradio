package rtp

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/rtp"
)

// Sender sends RTP packets (Opus audio)
type Sender struct {
	conn           *net.UDPConn
	ssrc           uint32        // Synchronization source identifier
	payloadType    uint8         // 111 = Opus
	sequenceNumber uint16
	timestamp      uint32
	sampleRate     uint32        // For timestamp calculation
	mu             sync.Mutex
}

// SenderConfig configures RTP sender
type SenderConfig struct {
	LocalPort   int
	PayloadType uint8  // 111 = Opus (RFC 7587)
	SSRC        uint32 // Random identifier for this stream
	SampleRate  uint32 // 48000 for Opus
}

// NewSender creates a new RTP sender
func NewSender(config SenderConfig) (*Sender, error) {
	// Default values
	if config.PayloadType == 0 {
		config.PayloadType = 111 // Opus
	}
	if config.SSRC == 0 {
		config.SSRC = uint32(time.Now().Unix()) // Random SSRC
	}
	if config.SampleRate == 0 {
		config.SampleRate = 48000 // 48kHz for Opus
	}

	// Create UDP socket
	addr := &net.UDPAddr{
		IP:   net.IPv6zero,
		Port: config.LocalPort,
	}

	conn, err := net.ListenUDP("udp6", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket: %w", err)
	}

	return &Sender{
		conn:           conn,
		ssrc:           config.SSRC,
		payloadType:    config.PayloadType,
		sequenceNumber: 0,
		timestamp:      0,
		sampleRate:     config.SampleRate,
	}, nil
}

// SendOpus sends Opus-encoded audio as RTP packet
func (s *Sender) SendOpus(opusData []byte, destination *net.UDPAddr) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create RTP packet
	packet := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			Padding:        false,
			Extension:      false,
			Marker:         false,
			PayloadType:    s.payloadType,
			SequenceNumber: s.sequenceNumber,
			Timestamp:      s.timestamp,
			SSRC:           s.ssrc,
		},
		Payload: opusData,
	}

	// Marshal to bytes
	data, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal RTP packet: %w", err)
	}

	// Send via UDP
	_, err = s.conn.WriteToUDP(data, destination)
	if err != nil {
		return fmt.Errorf("failed to send RTP packet: %w", err)
	}

	// Increment sequence number
	s.sequenceNumber++

	// Increment timestamp (20ms frame at 48kHz = 960 samples)
	s.timestamp += 960

	return nil
}

// SendToMultiple sends RTP packet to multiple destinations (fan-out)
func (s *Sender) SendToMultiple(opusData []byte, destinations []*net.UDPAddr) error {
	s.mu.Lock()

	// Create RTP packet (once)
	packet := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			Padding:        false,
			Extension:      false,
			Marker:         false,
			PayloadType:    s.payloadType,
			SequenceNumber: s.sequenceNumber,
			Timestamp:      s.timestamp,
			SSRC:           s.ssrc,
		},
		Payload: opusData,
	}

	data, err := packet.Marshal()
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to marshal RTP packet: %w", err)
	}

	// Update sequence/timestamp
	s.sequenceNumber++
	s.timestamp += 960

	s.mu.Unlock()

	// Send to all destinations (in parallel)
	var wg sync.WaitGroup
	errors := make(chan error, len(destinations))

	for _, dest := range destinations {
		wg.Add(1)
		go func(addr *net.UDPAddr) {
			defer wg.Done()
			_, err := s.conn.WriteToUDP(data, addr)
			if err != nil {
				errors <- err
			}
		}(dest)
	}

	wg.Wait()
	close(errors)

	// Return first error if any
	for err := range errors {
		return err
	}

	return nil
}

// Close closes the sender
func (s *Sender) Close() error {
	return s.conn.Close()
}

// GetStats returns sender statistics
func (s *Sender) GetStats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	return Stats{
		PacketsSent:    uint64(s.sequenceNumber),
		SSRC:           s.ssrc,
		CurrentTimestamp: s.timestamp,
	}
}

type Stats struct {
	PacketsSent      uint64
	SSRC             uint32
	CurrentTimestamp uint32
}
