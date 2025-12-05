package listener

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/meshradio/meshradio/pkg/audio"
	"github.com/meshradio/meshradio/pkg/emergency"
	"github.com/meshradio/meshradio/pkg/network"
	"github.com/meshradio/meshradio/pkg/protocol"
)

// Listener receives and plays audio streams
type Listener struct {
	callsign    string
	localIPv6   net.IP
	localPort   int
	targetIPv6  net.IP
	targetPort  int
	group       string  // Multicast group (e.g., "emergency", "community")
	ssmSource   net.IP  // SSM source (nil = regular multicast)
	transport   *network.Transport
	audioOut    *audio.OutputStream
	codec       audio.Codec
	config      audio.StreamConfig
	running     bool
	mu          sync.Mutex
	stopChan    chan struct{}

	// Subscription state
	subscribed      bool
	lastHeartbeat   time.Time

	// Stats
	packetsReceived uint64
	lastSeqNum      uint8
	stationCallsign string

	// Emergency handling (Layer 5)
	emergencySettings emergency.EmergencySettings
	lastPriority      uint8
}

// Config holds listener configuration
type Config struct {
	Callsign    string
	LocalIPv6   net.IP
	LocalPort   int
	TargetIPv6  net.IP
	TargetPort  int
	Group       string  // Multicast group (e.g., "emergency", "community")
	SSMSource   net.IP  // SSM source (nil = regular multicast, receives from all)
	AudioConfig audio.StreamConfig
}

// New creates a new listener
func New(cfg Config) (*Listener, error) {
	transport, err := network.NewTransport(cfg.LocalPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	audioOut := audio.NewOutputStream(cfg.AudioConfig)

	// Create Opus codec for decompression
	codec, err := audio.NewOpusCodec(
		cfg.AudioConfig.SampleRate,
		cfg.AudioConfig.Channels,
		cfg.AudioConfig.FrameSize,
		cfg.AudioConfig.Bitrate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus codec: %w", err)
	}

	// Default group if not specified
	group := cfg.Group
	if group == "" {
		group = "default"
	}

	return &Listener{
		callsign:          cfg.Callsign,
		localIPv6:         cfg.LocalIPv6,
		localPort:         cfg.LocalPort,
		targetIPv6:        cfg.TargetIPv6,
		targetPort:        cfg.TargetPort,
		group:             group,
		ssmSource:         cfg.SSMSource,
		transport:         transport,
		audioOut:          audioOut,
		codec:             codec,
		config:            cfg.AudioConfig,
		stopChan:          make(chan struct{}),
		emergencySettings: emergency.DefaultSettings(),
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

	// Send SUBSCRIBE packet to broadcaster
	if err := l.subscribe(); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Start receive loop
	go l.receiveLoop()

	// Start heartbeat loop
	go l.heartbeatLoop()

	fmt.Printf("Subscribed to %s:%d\n", l.targetIPv6.String(), l.targetPort)

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
			// Decode audio in separate goroutine to avoid blocking receive loop
			// This prevents packets from queuing up in the network buffer
			go l.handleAudioPacket(packet)
		case protocol.PacketTypeBeacon:
			l.handleBeacon(packet)
		case protocol.PacketTypeMetadata:
			l.handleMetadata(packet)
		}
	}
}

// handleAudioPacket processes an audio packet
func (l *Listener) handleAudioPacket(packet *protocol.Packet) {
	// Thread-safe increment since this runs in goroutines
	count := atomic.AddUint64(&l.packetsReceived, 1)

	// Get priority from packet (Layer 5: Emergency)
	priority := packet.GetPriority()

	// Check for priority change (emergency broadcast)
	if priority != l.lastPriority {
		l.handlePriorityChange(packet, priority)
		l.lastPriority = priority
	}

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

	// Log periodically with priority indicator
	if count%50 == 0 {
		priorityStr := ""
		if priority > 0 {
			p := emergency.Priority(priority)
			priorityStr = fmt.Sprintf(" [%s]", p.String())
		}
		fmt.Printf("Received: packets=%d, seq=%d, from=%s%s\n",
			count, packet.SequenceNum, packet.GetCallsign(), priorityStr)
	}

	l.lastSeqNum = packet.SequenceNum
}

// handlePriorityChange handles priority level changes
func (l *Listener) handlePriorityChange(packet *protocol.Packet, priority uint8) {
	p := emergency.Priority(priority)

	// Only log significant priority changes
	if priority < uint8(emergency.PriorityHigh) {
		return
	}

	sourceIPv6 := protocol.BytesToIPv6(packet.SourceIPv6)
	callsign := packet.GetCallsign()

	switch {
	case priority >= uint8(emergency.PriorityCritical):
		fmt.Printf("\n🚨 CRITICAL EMERGENCY BROADCAST from %s (%s)\n",
			callsign, sourceIPv6)
		fmt.Printf("   Priority: %s | Group: %s\n\n", p.String(), l.group)

	case priority >= uint8(emergency.PriorityEmergency):
		fmt.Printf("\n⚠️  EMERGENCY BROADCAST from %s (%s)\n",
			callsign, sourceIPv6)
		fmt.Printf("   Priority: %s | Group: %s\n\n", p.String(), l.group)

	case priority >= uint8(emergency.PriorityHigh):
		fmt.Printf("\n📢 High priority broadcast from %s (%s)\n",
			callsign, sourceIPv6)
		fmt.Printf("   Priority: %s | Group: %s\n\n", p.String(), l.group)
	}
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
	return atomic.LoadUint64(&l.packetsReceived), l.lastSeqNum, l.stationCallsign
}

// IsRunning returns whether the listener is running
func (l *Listener) IsRunning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}

// subscribe sends a SUBSCRIBE packet to the broadcaster
func (l *Listener) subscribe() error {
	var ipv6Bytes [16]byte
	copy(ipv6Bytes[:], l.localIPv6.To16())

	var callsignBytes [16]byte
	copy(callsignBytes[:], []byte(l.callsign))

	// Convert group name to bytes
	groupBytes := protocol.StringToGroup(l.group)

	// Convert SSM source to bytes (all zeros if regular multicast)
	var ssmSourceBytes [16]byte
	if l.ssmSource != nil {
		copy(ssmSourceBytes[:], l.ssmSource.To16())
	}

	subPayload := &protocol.SubscribePayload{
		ListenerIPv6: ipv6Bytes,
		ListenerPort: uint16(l.localPort),
		Callsign:     callsignBytes,
		Group:        groupBytes,
		SSMSource:    ssmSourceBytes,
	}

	packet := protocol.NewPacket(
		protocol.PacketTypeSubscribe,
		ipv6Bytes,
		l.callsign,
		protocol.MarshalSubscribe(subPayload),
	)

	err := l.transport.Send(packet, l.targetIPv6, l.targetPort)
	if err != nil {
		return fmt.Errorf("failed to send subscribe: %w", err)
	}

	l.subscribed = true
	l.lastHeartbeat = time.Now()

	multicastType := "Regular multicast"
	if l.ssmSource != nil {
		multicastType = fmt.Sprintf("SSM (source=%s)", l.ssmSource)
	}
	fmt.Printf("Sent SUBSCRIBE to %s:%d [%s] group='%s'\n",
		l.targetIPv6.String(), l.targetPort, multicastType, l.group)

	return nil
}

// heartbeatLoop sends periodic heartbeats to broadcaster
func (l *Listener) heartbeatLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !l.subscribed {
				continue
			}

			var ipv6Bytes [16]byte
			copy(ipv6Bytes[:], l.localIPv6.To16())

			hbPayload := &protocol.HeartbeatPayload{
				ListenerIPv6: ipv6Bytes,
				Timestamp:    uint64(time.Now().Unix()),
			}

			packet := protocol.NewPacket(
				protocol.PacketTypeHeartbeat,
				ipv6Bytes,
				l.callsign,
				protocol.MarshalHeartbeat(hbPayload),
			)

			err := l.transport.Send(packet, l.targetIPv6, l.targetPort)
			if err == nil {
				l.lastHeartbeat = time.Now()
			}

		case <-l.stopChan:
			return
		}
	}
}
