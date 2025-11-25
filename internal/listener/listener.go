package listener

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/meshradio/meshradio/pkg/audio"
	"github.com/meshradio/meshradio/pkg/network"
	"github.com/meshradio/meshradio/pkg/protocol"
)

// Listener receives and plays audio streams
type Listener struct {
	targetIPv6  net.IP
	targetPort  int
	transport   *network.Transport
	audioOut    *audio.OutputStream
	codec       audio.Codec
	config      audio.StreamConfig
	running     bool
	mu          sync.Mutex
	stopChan    chan struct{}

	// Stats
	packetsReceived uint64
	lastSeqNum      uint8
	stationCallsign string
}

// Config holds listener configuration
type Config struct {
	TargetIPv6  net.IP
	TargetPort  int
	LocalPort   int
	AudioConfig audio.StreamConfig
}

// New creates a new listener
func New(cfg Config) (*Listener, error) {
	transport, err := network.NewTransport(cfg.LocalPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	audioOut := audio.NewOutputStream(cfg.AudioConfig)
	codec := audio.NewDummyCodec(cfg.AudioConfig.FrameSize)

	return &Listener{
		targetIPv6: cfg.TargetIPv6,
		targetPort: cfg.TargetPort,
		transport:  transport,
		audioOut:   audioOut,
		codec:      codec,
		config:     cfg.AudioConfig,
		stopChan:   make(chan struct{}),
	}, nil
}

// Start begins listening
func (l *Listener) Start() error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return fmt.Errorf("listener already running")
	}
	l.running = true
	l.mu.Unlock()

	// Start transport
	if err := l.transport.Start(); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	// Start audio output
	if err := l.audioOut.Start(); err != nil {
		return fmt.Errorf("failed to start audio output: %w", err)
	}

	// Start receive loop
	go l.receiveLoop()

	fmt.Printf("Listening to %s:%d\n", l.targetIPv6.String(), l.targetPort)

	return nil
}

// Stop stops listening
func (l *Listener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return nil
	}

	l.running = false
	close(l.stopChan)

	l.audioOut.Stop()
	l.transport.Stop()

	return nil
}

// receiveLoop continuously receives and plays audio packets
func (l *Listener) receiveLoop() {
	fmt.Println("Receive loop started")

	for {
		select {
		case <-l.stopChan:
			return
		default:
		}

		// Receive packet
		packet, err := l.transport.Receive()
		if err != nil {
			if l.running {
				time.Sleep(100 * time.Millisecond)
			}
			continue
		}

		// Handle different packet types
		switch packet.Type {
		case protocol.PacketTypeAudio:
			l.handleAudioPacket(packet)
		case protocol.PacketTypeBeacon:
			l.handleBeacon(packet)
		case protocol.PacketTypeMetadata:
			l.handleMetadata(packet)
		}
	}
}

// handleAudioPacket processes an audio packet
func (l *Listener) handleAudioPacket(packet *protocol.Packet) {
	l.packetsReceived++

	// Parse audio payload
	audioPacket, err := protocol.UnmarshalAudioPayload(packet.Payload)
	if err != nil {
		fmt.Printf("Failed to unmarshal audio: %v\n", err)
		return
	}

	// Decode audio
	pcm, err := l.codec.Decode(audioPacket.AudioData)
	if err != nil {
		fmt.Printf("Failed to decode audio: %v\n", err)
		return
	}

	// Play audio
	l.audioOut.Write(pcm)

	// Log periodically
	if l.packetsReceived%50 == 0 {
		fmt.Printf("Received: packets=%d, seq=%d, from=%s\n",
			l.packetsReceived, packet.SequenceNum, packet.GetCallsign())
	}

	l.lastSeqNum = packet.SequenceNum
}

// handleBeacon processes a beacon packet
func (l *Listener) handleBeacon(packet *protocol.Packet) {
	callsign := packet.GetCallsign()
	if l.stationCallsign == "" {
		l.stationCallsign = callsign
		fmt.Printf("Connected to station: %s\n", callsign)
	}
}

// handleMetadata processes a metadata packet
func (l *Listener) handleMetadata(packet *protocol.Packet) {
	fmt.Printf("Metadata from %s: %s\n", packet.GetCallsign(), string(packet.Payload))
}

// GetStats returns listener statistics
func (l *Listener) GetStats() (packetsReceived uint64, lastSeq uint8, station string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.packetsReceived, l.lastSeqNum, l.stationCallsign
}

// IsRunning returns whether the listener is running
func (l *Listener) IsRunning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}
