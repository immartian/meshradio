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

// ListenerConn represents a connected listener
type ListenerConn struct {
	IPv6        net.IP
	Port        uint16
	Callsign    string
	ConnectedAt time.Time
	LastSeen    time.Time
}

// Broadcaster streams audio to subscribed listeners
type Broadcaster struct {
	callsign   string
	ipv6       net.IP
	port       int
	transport  *network.Transport
	audioIn    *audio.InputStream
	codec      audio.Codec
	config     audio.StreamConfig
	running    bool
	mu         sync.Mutex
	seqNum     uint8
	stopChan   chan struct{}

	// Subscription-based listener management
	listeners    map[string]*ListenerConn // key: "ipv6:port"
	listenersMux sync.RWMutex
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
		port:      cfg.Port,
		transport: transport,
		audioIn:   audioIn,
		codec:     codec,
		config:    cfg.AudioConfig,
		stopChan:  make(chan struct{}),
		listeners: make(map[string]*ListenerConn),
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

	// Handle incoming subscriptions and heartbeats
	go b.subscriptionLoop()

	// Monitor listener timeouts
	go b.heartbeatMonitor()

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

// broadcastLoop continuously broadcasts audio to subscribed listeners
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

		// Send to all subscribed listeners (unicast to each)
		b.listenersMux.RLock()
		listenerCount := len(b.listeners)
		for _, listener := range b.listeners {
			err = b.transport.Send(packet, listener.IPv6, int(listener.Port))
			if err != nil && b.seqNum%100 == 0 {
				fmt.Printf("Send error to %s: %v\n", listener.Callsign, err)
			}
		}
		b.listenersMux.RUnlock()

		// Log periodically
		if b.seqNum%50 == 0 { // Log every 50 frames (~1 second)
			fmt.Printf("Broadcasting: seq=%d, size=%d bytes to %d listeners\n",
				packet.SequenceNum, len(encoded), listenerCount)
		}
	}
}

// subscriptionLoop handles incoming SUBSCRIBE and HEARTBEAT packets
func (b *Broadcaster) subscriptionLoop() {
	for {
		select {
		case <-b.stopChan:
			return
		default:
		}

		// Receive packet from transport
		packet, err := b.transport.Receive()
		if err != nil {
			continue
		}

		switch packet.Type {
		case protocol.PacketTypeSubscribe:
			b.handleSubscribe(packet)
		case protocol.PacketTypeHeartbeat:
			b.handleHeartbeat(packet)
		}
	}
}

// handleSubscribe processes a subscription request
func (b *Broadcaster) handleSubscribe(packet *protocol.Packet) {
	sub, err := protocol.UnmarshalSubscribe(packet.Payload)
	if err != nil {
		fmt.Printf("Invalid subscribe packet: %v\n", err)
		return
	}

	listenerIP := protocol.BytesToIPv6(sub.ListenerIPv6)
	listenerKey := fmt.Sprintf("%s:%d", listenerIP.String(), sub.ListenerPort)

	b.listenersMux.Lock()
	b.listeners[listenerKey] = &ListenerConn{
		IPv6:        listenerIP,
		Port:        sub.ListenerPort,
		Callsign:    protocol.GetCallsignString(sub.Callsign),
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
	b.listenersMux.Unlock()

	fmt.Printf("New listener: %s (%s)\n", protocol.GetCallsignString(sub.Callsign), listenerIP.String())
}

// handleHeartbeat processes a heartbeat from listener
func (b *Broadcaster) handleHeartbeat(packet *protocol.Packet) {
	hb, err := protocol.UnmarshalHeartbeat(packet.Payload)
	if err != nil {
		return
	}

	listenerIP := protocol.BytesToIPv6(hb.ListenerIPv6)

	b.listenersMux.Lock()
	for key, listener := range b.listeners {
		if listener.IPv6.Equal(listenerIP) {
			listener.LastSeen = time.Now()
			b.listeners[key] = listener
			break
		}
	}
	b.listenersMux.Unlock()
}

// heartbeatMonitor removes listeners that haven't sent heartbeat
func (b *Broadcaster) heartbeatMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.listenersMux.Lock()
			for key, listener := range b.listeners {
				if time.Since(listener.LastSeen) > 15*time.Second {
					fmt.Printf("Listener timeout: %s\n", listener.Callsign)
					delete(b.listeners, key)
				}
			}
			b.listenersMux.Unlock()

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

// GetListenerCount returns the number of connected listeners
func (b *Broadcaster) GetListenerCount() int {
	b.listenersMux.RLock()
	defer b.listenersMux.RUnlock()
	return len(b.listeners)
}

// GetListeners returns a snapshot of current listeners
func (b *Broadcaster) GetListeners() []ListenerConn {
	b.listenersMux.RLock()
	defer b.listenersMux.RUnlock()

	listeners := make([]ListenerConn, 0, len(b.listeners))
	for _, l := range b.listeners {
		listeners = append(listeners, *l)
	}
	return listeners
}
