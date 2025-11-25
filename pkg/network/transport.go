package network

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/meshradio/meshradio/pkg/protocol"
)

// Transport handles UDP-based packet transmission
type Transport struct {
	conn       *net.UDPConn
	localAddr  *net.UDPAddr
	remoteAddr *net.UDPAddr
	running    bool
	mu         sync.Mutex

	// Packet receive channel
	packets chan *protocol.Packet
}

// NewTransport creates a new transport layer
func NewTransport(localPort int) (*Transport, error) {
	addr := &net.UDPAddr{
		IP:   net.IPv6zero,
		Port: localPort,
	}

	conn, err := net.ListenUDP("udp6", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket: %w", err)
	}

	return &Transport{
		conn:      conn,
		localAddr: addr,
		packets:   make(chan *protocol.Packet, 100),
	}, nil
}

// Start begins listening for packets
func (t *Transport) Start() error {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return fmt.Errorf("transport already running")
	}
	t.running = true
	t.mu.Unlock()

	go t.receiveLoop()
	return nil
}

// Stop stops the transport
func (t *Transport) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	t.running = false
	return t.conn.Close()
}

// Send sends a packet to the specified IPv6 address
func (t *Transport) Send(packet *protocol.Packet, targetIPv6 net.IP, port int) error {
	data, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	addr := &net.UDPAddr{
		IP:   targetIPv6,
		Port: port,
	}

	_, err = t.conn.WriteToUDP(data, addr)
	if err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}

	return nil
}

// Receive returns the next received packet
func (t *Transport) Receive() (*protocol.Packet, error) {
	packet, ok := <-t.packets
	if !ok {
		return nil, fmt.Errorf("transport closed")
	}
	return packet, nil
}

// SetRemote sets the remote address for sending
func (t *Transport) SetRemote(ipv6 net.IP, port int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.remoteAddr = &net.UDPAddr{
		IP:   ipv6,
		Port: port,
	}
}

// receiveLoop continuously receives packets
func (t *Transport) receiveLoop() {
	buffer := make([]byte, 65535) // Max UDP packet size

	for t.running {
		// Set read deadline to allow checking running status
		t.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, _, err := t.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout, check running status
			}
			if t.running {
				fmt.Printf("Error reading UDP: %v\n", err)
			}
			continue
		}

		// Parse packet
		packet, err := protocol.Unmarshal(buffer[:n])
		if err != nil {
			fmt.Printf("Error unmarshaling packet: %v\n", err)
			continue
		}

		// Queue packet
		select {
		case t.packets <- packet:
		default:
			// Channel full, drop packet
		}
	}

	close(t.packets)
}

// LocalAddr returns the local address
func (t *Transport) LocalAddr() *net.UDPAddr {
	return t.localAddr
}
