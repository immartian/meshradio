package multicast

import (
	"net"
	"time"
)

// Subscriber represents a listener subscribed to a multicast group
type Subscriber struct {
	IPv6      net.IP    // Subscriber's IPv6 address
	Port      int       // RTP port
	Callsign  string    // Station callsign
	LastSeen  time.Time // Last heartbeat received
	SSMSource net.IP    // nil = regular multicast, non-nil = SSM (only receive from this source)
}

// Broadcaster represents a broadcaster in a multicast group
type Broadcaster struct {
	IPv6     net.IP    // Broadcaster's IPv6 address
	Port     int       // RTP port
	Callsign string    // Station callsign
	LastSeen time.Time // Last heartbeat
}

// Group represents a multicast group
type Group struct {
	Name         string                    // Group name (e.g., "emergency", "community")
	Subscribers  map[string]*Subscriber    // Key: IPv6:port
	Broadcasters map[string]*Broadcaster   // Key: IPv6
}

// SubscribeRequest represents a subscription request
type SubscribeRequest struct {
	Group      string     // Group name
	Subscriber *Subscriber
}

// UnsubscribeRequest represents an unsubscribe request
type UnsubscribeRequest struct {
	Group string // Group name
	IPv6  net.IP // Subscriber IPv6
	Port  int    // Subscriber port
}

// GetKey returns a unique key for a subscriber (IPv6:port)
func (s *Subscriber) GetKey() string {
	return net.JoinHostPort(s.IPv6.String(), string(rune(s.Port)))
}

// GetKey returns a unique key for a broadcaster (IPv6)
func (b *Broadcaster) GetKey() string {
	return b.IPv6.String()
}

// IsRegularMulticast returns true if this is a regular multicast subscription
func (s *Subscriber) IsRegularMulticast() bool {
	return s.SSMSource == nil
}

// IsSSM returns true if this is an SSM subscription
func (s *Subscriber) IsSSM() bool {
	return s.SSMSource != nil
}

// MatchesSource returns true if subscriber wants packets from this source
func (s *Subscriber) MatchesSource(source net.IP) bool {
	// Regular multicast: accept all sources
	if s.IsRegularMulticast() {
		return true
	}

	// SSM: only accept from specified source
	return s.SSMSource.Equal(source)
}
