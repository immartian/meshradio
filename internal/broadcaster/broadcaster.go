package broadcaster

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/meshradio/meshradio/pkg/audio"
	"github.com/meshradio/meshradio/pkg/network"
	"github.com/meshradio/meshradio/pkg/protocol"
)

// Broadcaster streams audio to the network
type Broadcaster struct {
	callsign   string
	ipv6       net.IP
	transport  *network.Transport
	audioIn    *audio.InputStream
	codec      audio.Codec
	config     audio.StreamConfig
	running    bool
	mu         sync.Mutex
	seqNum     uint8
	stopChan   chan struct{}
}

// Config holds broadcaster configuration
type Config struct {
	Callsign   string
	IPv6       net.IP
	Port       int
	AudioConfig audio.StreamConfig
}

// New creates a new broadcaster
func New(cfg Config) (*Broadcaster, error) {
	transport, err := network.NewTransport(cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	audioIn := audio.NewInputStream(cfg.AudioConfig)
	codec := audio.NewDummyCodec(cfg.AudioConfig.FrameSize)

	return &Broadcaster{
		callsign:  cfg.Callsign,
		ipv6:      cfg.IPv6,
		transport: transport,
		audioIn:   audioIn,
		codec:     codec,
		config:    cfg.AudioConfig,
		stopChan:  make(chan struct{}),
	}, nil
}

// Start begins broadcasting
func (b *Broadcaster) Start() error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return fmt.Errorf("broadcaster already running")
	}
	b.running = true
	b.mu.Unlock()

	// Start transport
	if err := b.transport.Start(); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	// Start audio input
	if err := b.audioIn.Start(); err != nil {
		return fmt.Errorf("failed to start audio input: %w", err)
	}

	// Start broadcast loop
	go b.broadcastLoop()

	// Send periodic beacons
	go b.beaconLoop()

	return nil
}

// Stop stops broadcasting
func (b *Broadcaster) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil
	}

	b.running = false
	close(b.stopChan)

	b.audioIn.Stop()
	b.transport.Stop()

	return nil
}

// broadcastLoop continuously broadcasts audio
func (b *Broadcaster) broadcastLoop() {
	var ipv6Bytes [16]byte
	copy(ipv6Bytes[:], b.ipv6.To16())

	for {
		select {
		case <-b.stopChan:
			return
		default:
		}

		// Read audio frame
		pcm, err := b.audioIn.Read()
		if err != nil {
			continue
		}

		// Encode audio
		encoded, err := b.codec.Encode(pcm)
		if err != nil {
			fmt.Printf("Encode error: %v\n", err)
			continue
		}

		// Create audio packet payload
		audioPayload := protocol.MarshalAudioPayload(&protocol.AudioPacket{
			CodecType:      protocol.CodecOpus,
			SampleRate:     uint8(b.config.SampleRate / 1000),
			Channels:       uint8(b.config.Channels),
			Bitrate:        uint8(b.config.Bitrate / 1000),
			FrameTimestamp: uint32(time.Now().UnixMilli()),
			AudioData:      encoded,
		})

		// Create protocol packet
		packet := protocol.NewPacket(
			protocol.PacketTypeAudio,
			ipv6Bytes,
			b.callsign,
			audioPayload,
		)
		packet.SequenceNum = b.seqNum
		b.seqNum++

		// Broadcast to multicast group (for MVP, we'll just prepare the packet)
		// In production, this would send to listeners
		_ = packet

		// For MVP, we'll log that we're broadcasting
		if b.seqNum%50 == 0 { // Log every 50 frames (~1 second)
			fmt.Printf("Broadcasting: seq=%d, size=%d bytes\n", packet.SequenceNum, len(encoded))
		}
	}
}

// beaconLoop sends periodic station beacons
func (b *Broadcaster) beaconLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var ipv6Bytes [16]byte
	copy(ipv6Bytes[:], b.ipv6.To16())

	for {
		select {
		case <-ticker.C:
			// Send beacon
			beacon := protocol.NewPacket(
				protocol.PacketTypeBeacon,
				ipv6Bytes,
				b.callsign,
				[]byte("{}"), // Empty JSON for MVP
			)

			fmt.Printf("Sending beacon: %s at %s\n", b.callsign, b.ipv6.String())
			_ = beacon

		case <-b.stopChan:
			return
		}
	}
}

// GetIPv6 returns the broadcaster's IPv6 address
func (b *Broadcaster) GetIPv6() net.IP {
	return b.ipv6
}

// GetCallsign returns the broadcaster's callsign
func (b *Broadcaster) GetCallsign() string {
	return b.callsign
}

// IsRunning returns whether the broadcaster is running
func (b *Broadcaster) IsRunning() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.running
}
