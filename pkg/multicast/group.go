package multicast

import (
	"fmt"
	"net"
	"time"
)

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
	key := net.JoinHostPort(sub.IPv6.String(), fmt.Sprintf("%d", sub.Port))
	g.Subscribers[key] = sub
}

// RemoveSubscriber removes a subscriber from the group
func (g *Group) RemoveSubscriber(ipv6 net.IP, port int) {
	key := net.JoinHostPort(ipv6.String(), fmt.Sprintf("%d", port))
	delete(g.Subscribers, key)
}

// GetSubscriber retrieves a subscriber by IPv6 and port
func (g *Group) GetSubscriber(ipv6 net.IP, port int) *Subscriber {
	key := net.JoinHostPort(ipv6.String(), fmt.Sprintf("%d", port))
	return g.Subscribers[key]
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

// AddBroadcaster adds a broadcaster to the group
func (g *Group) AddBroadcaster(broadcaster *Broadcaster) {
	key := broadcaster.IPv6.String()
	g.Broadcasters[key] = broadcaster
}

// RemoveBroadcaster removes a broadcaster from the group
func (g *Group) RemoveBroadcaster(ipv6 net.IP) {
	key := ipv6.String()
	delete(g.Broadcasters, key)
}

// GetBroadcaster retrieves a broadcaster by IPv6
func (g *Group) GetBroadcaster(ipv6 net.IP) *Broadcaster {
	key := ipv6.String()
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
		if now.Sub(sub.LastSeen) > timeout {
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
