package multicast

import (
	"fmt"
	"net"
	"time"
)

// makeSubscriberKey creates a consistent key for subscriber lookups
// Uses hex representation of IPv6 bytes to avoid string format inconsistencies
// (e.g., "::1" vs "0:0:0:0:0:0:0:1" both represent the same address)
func makeSubscriberKey(ipv6 net.IP, port int) string {
	// Convert to 16-byte canonical form and use hex to ensure consistency
	ipBytes := ipv6.To16()
	key := fmt.Sprintf("%x:%d", ipBytes, port)
	fmt.Printf("DEBUG: makeSubscriberKey: IP=%s, len(ipBytes)=%d, first4=%x, key=%s\n",
		ipv6.String(), len(ipBytes), ipBytes[:4], key)
	return key
}

// NewGroup creates a new multicast group
func NewGroup(name string) *Group {
	return &Group{
		Name:         name,
		Subscribers:  make(map[string]*Subscriber),
		Broadcasters: make(map[string]*Broadcaster),
	}
}

// AddSubscriber adds a subscriber to the group
func (g *Group) AddSubscriber(sub *Subscriber) {
	key := makeSubscriberKey(sub.IPv6, sub.Port)
	g.Subscribers[key] = sub
	fmt.Printf("DEBUG: AddSubscriber key='%s' for IP=%s port=%d\n", key, sub.IPv6.String(), sub.Port)
}

// RemoveSubscriber removes a subscriber from the group
func (g *Group) RemoveSubscriber(ipv6 net.IP, port int) {
	key := makeSubscriberKey(ipv6, port)
	delete(g.Subscribers, key)
	fmt.Printf("DEBUG: RemoveSubscriber key='%s' for IP=%s port=%d\n", key, ipv6.String(), port)
}

// GetSubscriber retrieves a subscriber by IPv6 and port
func (g *Group) GetSubscriber(ipv6 net.IP, port int) *Subscriber {
	key := makeSubscriberKey(ipv6, port)
	sub := g.Subscribers[key]
	fmt.Printf("DEBUG: GetSubscriber key='%s' for IP=%s port=%d, found=%v\n", key, ipv6.String(), port, sub != nil)
	return sub
}

// GetSubscribers returns all subscribers in the group
func (g *Group) GetSubscribers() []*Subscriber {
	subs := make([]*Subscriber, 0, len(g.Subscribers))
	for _, sub := range g.Subscribers {
		subs = append(subs, sub)
	}
	return subs
}

// GetSubscribersForSource returns subscribers that want packets from this source
func (g *Group) GetSubscribersForSource(source net.IP) []*Subscriber {
	subs := make([]*Subscriber, 0)
	for _, sub := range g.Subscribers {
		if sub.MatchesSource(source) {
			subs = append(subs, sub)
		}
	}
	return subs
}

// makeBroadcasterKey creates a consistent key for broadcaster lookups
func makeBroadcasterKey(ipv6 net.IP) string {
	// Use hex representation to avoid IPv6 string format inconsistencies
	return fmt.Sprintf("%x", ipv6.To16())
}

// AddBroadcaster adds a broadcaster to the group
func (g *Group) AddBroadcaster(broadcaster *Broadcaster) {
	key := makeBroadcasterKey(broadcaster.IPv6)
	g.Broadcasters[key] = broadcaster
}

// RemoveBroadcaster removes a broadcaster from the group
func (g *Group) RemoveBroadcaster(ipv6 net.IP) {
	key := makeBroadcasterKey(ipv6)
	delete(g.Broadcasters, key)
}

// GetBroadcaster retrieves a broadcaster by IPv6
func (g *Group) GetBroadcaster(ipv6 net.IP) *Broadcaster {
	key := makeBroadcasterKey(ipv6)
	return g.Broadcasters[key]
}

// GetBroadcasters returns all broadcasters in the group
func (g *Group) GetBroadcasters() []*Broadcaster {
	broadcasters := make([]*Broadcaster, 0, len(g.Broadcasters))
	for _, b := range g.Broadcasters {
		broadcasters = append(broadcasters, b)
	}
	return broadcasters
}

// PruneStaleSubscribers removes subscribers that haven't sent heartbeat
func (g *Group) PruneStaleSubscribers(timeout time.Duration) int {
	count := 0
	now := time.Now()

	for key, sub := range g.Subscribers {
		age := now.Sub(sub.LastSeen)
		if age > timeout {
			fmt.Printf("DEBUG: Pruning subscriber %s (key=%s, age=%v > timeout=%v)\n",
				sub.Callsign, key, age, timeout)
			delete(g.Subscribers, key)
			count++
		}
	}

	return count
}

// PruneStaleBroadcasters removes broadcasters that haven't sent heartbeat
func (g *Group) PruneStaleBroadcasters(timeout time.Duration) int {
	count := 0
	now := time.Now()

	for key, b := range g.Broadcasters {
		if now.Sub(b.LastSeen) > timeout {
			delete(g.Broadcasters, key)
			count++
		}
	}

	return count
}

// SubscriberCount returns the number of subscribers
func (g *Group) SubscriberCount() int {
	return len(g.Subscribers)
}

// BroadcasterCount returns the number of broadcasters
func (g *Group) BroadcasterCount() int {
	return len(g.Broadcasters)
}
