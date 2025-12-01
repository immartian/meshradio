package rtp

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/rtp"
)

// Receiver receives RTP packets (Opus audio)
type Receiver struct {
	conn           *net.UDPConn
	packets        chan *rtp.Packet
	running        bool
	mu             sync.Mutex

	// Statistics
	packetsReceived uint64
	packetsLost     uint64
	lastSeqNum      uint16
	jitterBuffer    *JitterBuffer
}

// ReceiverConfig configures RTP receiver
type ReceiverConfig struct {
	LocalPort    int
	BufferSize   int  // Jitter buffer size (packets)
}

// NewReceiver creates a new RTP receiver
func NewReceiver(config ReceiverConfig) (*Receiver, error) {
	// Default buffer size
	if config.BufferSize == 0 {
		config.BufferSize = 50 // ~1 second at 20ms frames
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

	return &Receiver{
		conn:         conn,
		packets:      make(chan *rtp.Packet, 100),
		jitterBuffer: NewJitterBuffer(config.BufferSize),
	}, nil
}

// Start begins receiving RTP packets
func (r *Receiver) Start() error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("receiver already running")
	}
	r.running = true
	r.mu.Unlock()

	go r.receiveLoop()
	return nil
}

// Stop stops receiving
func (r *Receiver) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return nil
	}

	r.running = false
	return r.conn.Close()
}

// receiveLoop continuously receives RTP packets
func (r *Receiver) receiveLoop() {
	buffer := make([]byte, 1500) // MTU size

	for r.running {
		// Set read deadline
		r.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, _, err := r.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if r.running {
				fmt.Printf("RTP receive error: %v\n", err)
			}
			continue
		}

		// Parse RTP packet
		packet := &rtp.Packet{}
		err = packet.Unmarshal(buffer[:n])
		if err != nil {
			fmt.Printf("Failed to unmarshal RTP packet: %v\n", err)
			continue
		}

		// Update statistics
		r.updateStats(packet)

		// Add to jitter buffer
		r.jitterBuffer.Add(packet)

		// Try to send ordered packet
		if orderedPacket := r.jitterBuffer.Pop(); orderedPacket != nil {
			select {
			case r.packets <- orderedPacket:
			default:
				// Channel full, drop packet
			}
		}
	}

	close(r.packets)
}

// ReadOpus reads the next Opus frame
func (r *Receiver) ReadOpus() ([]byte, error) {
	packet, ok := <-r.packets
	if !ok {
		return nil, fmt.Errorf("receiver closed")
	}
	return packet.Payload, nil
}

// updateStats updates receiver statistics
func (r *Receiver) updateStats(packet *rtp.Packet) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.packetsReceived++

	// Detect packet loss (sequence number gaps)
	if r.packetsReceived > 1 {
		expected := r.lastSeqNum + 1
		if packet.SequenceNumber != expected {
			// Packets lost
			var lost uint16
			if packet.SequenceNumber > expected {
				lost = packet.SequenceNumber - expected
			} else {
				// Sequence number wrapped around
				lost = (65535 - expected) + packet.SequenceNumber + 1
			}
			r.packetsLost += uint64(lost)
		}
	}

	r.lastSeqNum = packet.SequenceNumber
}

// GetStats returns receiver statistics
func (r *Receiver) GetStats() ReceiverStats {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lossRate float64
	if r.packetsReceived > 0 {
		lossRate = float64(r.packetsLost) / float64(r.packetsReceived+r.packetsLost) * 100
	}

	return ReceiverStats{
		PacketsReceived: r.packetsReceived,
		PacketsLost:     r.packetsLost,
		PacketLossRate:  lossRate,
	}
}

type ReceiverStats struct {
	PacketsReceived uint64
	PacketsLost     uint64
	PacketLossRate  float64 // Percentage
}

// JitterBuffer simple jitter buffer for packet reordering
type JitterBuffer struct {
	buffer   []*rtp.Packet
	size     int
	mu       sync.Mutex
}

// NewJitterBuffer creates a new jitter buffer
func NewJitterBuffer(size int) *JitterBuffer {
	return &JitterBuffer{
		buffer: make([]*rtp.Packet, 0, size),
		size:   size,
	}
}

// Add adds a packet to the buffer
func (jb *JitterBuffer) Add(packet *rtp.Packet) {
	jb.mu.Lock()
	defer jb.mu.Unlock()

	jb.buffer = append(jb.buffer, packet)

	// Keep buffer from growing too large
	if len(jb.buffer) > jb.size {
		jb.buffer = jb.buffer[1:]
	}
}

// Pop returns the next packet in sequence order
func (jb *JitterBuffer) Pop() *rtp.Packet {
	jb.mu.Lock()
	defer jb.mu.Unlock()

	if len(jb.buffer) == 0 {
		return nil
	}

	// Find packet with lowest sequence number
	minIdx := 0
	minSeq := jb.buffer[0].SequenceNumber

	for i, pkt := range jb.buffer {
		if pkt.SequenceNumber < minSeq {
			minSeq = pkt.SequenceNumber
			minIdx = i
		}
	}

	// Remove and return
	packet := jb.buffer[minIdx]
	jb.buffer = append(jb.buffer[:minIdx], jb.buffer[minIdx+1:]...)

	return packet
}
