package broadcaster

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/meshradio/meshradio/pkg/audio"
	"github.com/meshradio/meshradio/pkg/emergency"
	"github.com/meshradio/meshradio/pkg/multicast"
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
	callsign    string
	ipv6        net.IP
	port        int
	group       string  // Multicast group name (e.g., "emergency", "community")
	priority    uint8   // Broadcast priority (0-3)
	transport   *network.Transport
	audioSource audio.AudioSource // Can be microphone, MP3 file, etc.
	codec       audio.Codec
	config      audio.StreamConfig
	running     bool
	mu          sync.Mutex
	seqNum      uint8
	stopChan    chan struct{}

	// Subscription manager (Layer 4: Multicast Overlay)
	subManager *multicast.SubscriptionManager

	// Channel registry (Layer 5: Emergency)
	channelRegistry *emergency.ChannelRegistry

	// Legacy listener tracking (deprecated - use subManager instead)
	listeners    map[string]*ListenerConn // key: "ipv6:port"
	listenersMux sync.RWMutex
}

// Config holds broadcaster configuration
type Config struct {
	Callsign    string
	IPv6        net.IP
	Port        int
	Group       string              // Multicast group (e.g., "emergency", "community")
	AudioConfig audio.StreamConfig
	AudioSource audio.AudioSource   // Optional: custom audio source (microphone, MP3, etc.). If nil, uses microphone.
}

// New creates a new broadcaster
func New(cfg Config) (*Broadcaster, error) {
	transport, err := network.NewTransport(cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Use provided audio source, or default to microphone
	var audioSource audio.AudioSource
	if cfg.AudioSource != nil {
		audioSource = cfg.AudioSource
	} else {
		audioSource = audio.NewMicrophoneSource(cfg.AudioConfig)
	}

	codec := audio.NewDummyCodec(cfg.AudioConfig.FrameSize)

	// Default group if not specified
	group := cfg.Group
	if group == "" {
		group = "default"
	}

	// Get priority for this channel/group
	channelRegistry := emergency.NewChannelRegistry()
	priority := uint8(emergency.PriorityNormal) // Default
	if ch, ok := channelRegistry.GetByGroup(group); ok {
		priority = uint8(ch.Priority)
	}

	return &Broadcaster{
		callsign:        cfg.Callsign,
		ipv6:            cfg.IPv6,
		port:            cfg.Port,
		group:           group,
		priority:        priority,
		transport:       transport,
		audioSource:     audioSource,
		codec:           codec,
		config:          cfg.AudioConfig,
		stopChan:        make(chan struct{}),
		subManager:      multicast.NewSubscriptionManager(),
		channelRegistry: channelRegistry,
		listeners:       make(map[string]*ListenerConn),
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

	// Start audio source
	if err := b.audioSource.Start(); err != nil {
		return fmt.Errorf("failed to start audio source: %w", err)
	}

	// Register this broadcaster with the subscription manager
	broadcaster := &multicast.Broadcaster{
		IPv6:     b.ipv6,
		Port:     b.port,
		Callsign: b.callsign,
		LastSeen: time.Now(),
	}
	b.subManager.RegisterBroadcaster(b.group, broadcaster)

	// Get channel info for logging
	priorityStr := "normal"
	if ch, ok := b.channelRegistry.GetByGroup(b.group); ok {
		priorityStr = ch.Priority.String()
	}
	fmt.Printf("Registered broadcaster in group '%s' with priority '%s'\n", b.group, priorityStr)

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

	b.audioSource.Stop()
	b.transport.Stop()

	return nil
}

// broadcastLoop continuously broadcasts audio to subscribed listeners
func (b *Broadcaster) broadcastLoop() {
	var ipv6Bytes [16]byte
	copy(ipv6Bytes[:], b.ipv6.To16())

	fmt.Println("Broadcast loop started")

	for {
		select {
		case <-b.stopChan:
			return
		default:
		}

		// Read audio frame (as int16 samples)
		samples, err := b.audioSource.Read()
		if err != nil {
			if b.seqNum%50 == 0 { // Log periodically
				fmt.Printf("Audio source read error: %v\n", err)
			}
			continue
		}

		// Convert int16 samples to bytes for codec
		pcm := make([]byte, len(samples)*2)
		for i, sample := range samples {
			pcm[i*2] = byte(sample)
			pcm[i*2+1] = byte(sample >> 8)
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
		packet.SetPriority(b.priority) // Set priority (Layer 5: Emergency)
		b.seqNum++

		// Get subscribers for this broadcaster (using multicast overlay)
		subscribers := b.subManager.GetSubscribersForSource(b.group, b.ipv6)

		// Send to all subscribed listeners (unicast fan-out)
		for _, sub := range subscribers {
			err = b.transport.Send(packet, sub.IPv6, sub.Port)
			if err != nil && b.seqNum%100 == 0 {
				fmt.Printf("Send error to %s: %v\n", sub.Callsign, err)
			}
		}

		listenerCount := len(subscribers)

		// Log periodically
		if b.seqNum%50 == 0 { // Log every 50 frames (~1 second)
			fmt.Printf("Broadcasting: seq=%d, size=%d bytes to %d listeners\n",
				packet.SequenceNum, len(encoded), listenerCount)
		}
	}
}

// subscriptionLoop handles incoming SUBSCRIBE and HEARTBEAT packets
func (b *Broadcaster) subscriptionLoop() {
	fmt.Println("Subscription loop started")
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

		fmt.Printf("Received packet type=%d from %s\n", packet.Type, protocol.BytesToIPv6(packet.SourceIPv6))

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
	callsign := protocol.GetCallsignString(sub.Callsign)

	// Extract group name (use broadcaster's group if not specified)
	group := protocol.GetGroupString(sub.Group)
	if group == "" {
		group = b.group
	}

	// Extract SSM source (nil = regular multicast)
	var ssmSource net.IP
	if !protocol.IsZeroIPv6(sub.SSMSource) {
		ssmSource = protocol.BytesToIPv6(sub.SSMSource)
	}

	// Create subscriber
	subscriber := &multicast.Subscriber{
		IPv6:      listenerIP,
		Port:      int(sub.ListenerPort),
		Callsign:  callsign,
		LastSeen:  time.Now(),
		SSMSource: ssmSource,
	}

	// Add to subscription manager
	b.subManager.Subscribe(multicast.SubscribeRequest{
		Group:      group,
		Subscriber: subscriber,
	})

	multicastType := "Regular"
	if subscriber.IsSSM() {
		multicastType = fmt.Sprintf("SSM (source=%s)", ssmSource)
	}
	fmt.Printf("New subscriber: %s (%s) [%s] to group '%s'\n",
		callsign, listenerIP, multicastType, group)

	// Legacy: Also update old listeners map for backward compatibility
	listenerKey := fmt.Sprintf("%s:%d", listenerIP.String(), sub.ListenerPort)
	b.listenersMux.Lock()
	b.listeners[listenerKey] = &ListenerConn{
		IPv6:        listenerIP,
		Port:        sub.ListenerPort,
		Callsign:    callsign,
		ConnectedAt: time.Now(),
		LastSeen:    time.Now(),
	}
	b.listenersMux.Unlock()
}

// handleHeartbeat processes a heartbeat from listener
func (b *Broadcaster) handleHeartbeat(packet *protocol.Packet) {
	hb, err := protocol.UnmarshalHeartbeat(packet.Payload)
	if err != nil {
		return
	}

	listenerIP := protocol.BytesToIPv6(hb.ListenerIPv6)

	// Update heartbeat in subscription manager for all groups
	// (listener might be subscribed to multiple groups)
	for _, group := range b.subManager.ListGroups() {
		subs := b.subManager.GetSubscribers(group)
		for _, sub := range subs {
			if sub.IPv6.Equal(listenerIP) {
				b.subManager.Heartbeat(group, listenerIP, sub.Port)
			}
		}
	}

	// Legacy: Also update old listeners map for backward compatibility
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
			// Prune stale subscribers using multicast overlay
			prunedSubs, prunedBroadcasters := b.subManager.PruneStale(15 * time.Second)
			if prunedSubs > 0 || prunedBroadcasters > 0 {
				fmt.Printf("Pruned %d stale subscriber(s), %d broadcaster(s)\n",
					prunedSubs, prunedBroadcasters)
			}

			// Legacy: Also prune old listeners map
			b.listenersMux.Lock()
			for key, listener := range b.listeners {
				if time.Since(listener.LastSeen) > 15*time.Second {
					fmt.Printf("Listener timeout (legacy): %s\n", listener.Callsign)
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
