package multicast

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// SubscriptionManager manages multicast group subscriptions
type SubscriptionManager struct {
	groups map[string]*Group // Key: group name
	mu     sync.RWMutex
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		groups: make(map[string]*Group),
	}
}

// Subscribe adds a subscriber to a group
func (sm *SubscriptionManager) Subscribe(req SubscribeRequest) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Create group if it doesn't exist
	if _, exists := sm.groups[req.Group]; !exists {
		sm.groups[req.Group] = NewGroup(req.Group)
	}

	group := sm.groups[req.Group]

	// Update LastSeen timestamp
	req.Subscriber.LastSeen = time.Now()

	// Add subscriber to group
	group.AddSubscriber(req.Subscriber)

	return nil
}

// Unsubscribe removes a subscriber from a group
func (sm *SubscriptionManager) Unsubscribe(req UnsubscribeRequest) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	group, exists := sm.groups[req.Group]
	if !exists {
		return fmt.Errorf("group not found: %s", req.Group)
	}

	group.RemoveSubscriber(req.IPv6, req.Port)

	// Remove group if empty
	if group.SubscriberCount() == 0 && group.BroadcasterCount() == 0 {
		fmt.Printf("DEBUG: Deleting empty group '%s' after Unsubscribe\n", req.Group)
		delete(sm.groups, req.Group)
	}

	return nil
}

// Heartbeat updates the LastSeen timestamp for a subscriber
func (sm *SubscriptionManager) Heartbeat(group string, ipv6 net.IP, port int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	g, exists := sm.groups[group]
	if !exists {
		return fmt.Errorf("group not found: %s", group)
	}

	sub := g.GetSubscriber(ipv6, port)
	if sub == nil {
		return fmt.Errorf("subscriber not found: %s:%d", ipv6, port)
	}

	oldLastSeen := sub.LastSeen
	sub.LastSeen = time.Now()
	fmt.Printf("DEBUG: Heartbeat updated %s (age was %v)\n", sub.Callsign, time.Since(oldLastSeen))
	return nil
}

// GetSubscribers returns all subscribers for a group
func (sm *SubscriptionManager) GetSubscribers(group string) []*Subscriber {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	g, exists := sm.groups[group]
	if !exists {
		return nil
	}

	return g.GetSubscribers()
}

// GetSubscribersForSource returns subscribers that want packets from this source
func (sm *SubscriptionManager) GetSubscribersForSource(group string, source net.IP) []*Subscriber {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	g, exists := sm.groups[group]
	if !exists {
		return nil
	}

	return g.GetSubscribersForSource(source)
}

// RegisterBroadcaster registers a broadcaster for a group
func (sm *SubscriptionManager) RegisterBroadcaster(group string, broadcaster *Broadcaster) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Create group if it doesn't exist
	if _, exists := sm.groups[group]; !exists {
		fmt.Printf("DEBUG: RegisterBroadcaster creating new group '%s'\n", group)
		sm.groups[group] = NewGroup(group)
	}

	g := sm.groups[group]

	// Update LastSeen timestamp
	broadcaster.LastSeen = time.Now()

	// Add broadcaster to group
	g.AddBroadcaster(broadcaster)

	return nil
}

// UnregisterBroadcaster removes a broadcaster from a group
func (sm *SubscriptionManager) UnregisterBroadcaster(group string, ipv6 net.IP) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	g, exists := sm.groups[group]
	if !exists {
		return fmt.Errorf("group not found: %s", group)
	}

	g.RemoveBroadcaster(ipv6)

	// Remove group if empty
	if g.SubscriberCount() == 0 && g.BroadcasterCount() == 0 {
		fmt.Printf("DEBUG: Deleting empty group '%s' after UnregisterBroadcaster\n", group)
		delete(sm.groups, group)
	}

	return nil
}

// GetBroadcasters returns all broadcasters for a group
func (sm *SubscriptionManager) GetBroadcasters(group string) []*Broadcaster {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	g, exists := sm.groups[group]
	if !exists {
		return nil
	}

	return g.GetBroadcasters()
}

// CreateGroup creates a new group
func (sm *SubscriptionManager) CreateGroup(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.groups[name]; exists {
		return fmt.Errorf("group already exists: %s", name)
	}

	sm.groups[name] = NewGroup(name)
	return nil
}

// DeleteGroup deletes a group
func (sm *SubscriptionManager) DeleteGroup(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.groups[name]; !exists {
		return fmt.Errorf("group not found: %s", name)
	}

	fmt.Printf("DEBUG: Deleting group '%s' (explicit DeleteGroup call)\n", name)
	delete(sm.groups, name)
	return nil
}

// ListGroups returns all group names
func (sm *SubscriptionManager) ListGroups() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.groups))
	for name := range sm.groups {
		names = append(names, name)
	}
	return names
}

// GetGroupInfo returns information about a group
func (sm *SubscriptionManager) GetGroupInfo(name string) (*Group, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	g, exists := sm.groups[name]
	if !exists {
		return nil, fmt.Errorf("group not found: %s", name)
	}

	return g, nil
}

// PruneStale removes stale subscribers and broadcasters from all groups
func (sm *SubscriptionManager) PruneStale(timeout time.Duration) (int, int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	totalSubs := 0
	totalBroadcasters := 0

	for _, group := range sm.groups {
		totalSubs += group.PruneStaleSubscribers(timeout)
		totalBroadcasters += group.PruneStaleBroadcasters(timeout)
	}

	return totalSubs, totalBroadcasters
}

// GetStats returns statistics about subscriptions
func (sm *SubscriptionManager) GetStats() Stats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := Stats{
		GroupCount: len(sm.groups),
		Groups:     make(map[string]GroupStats),
	}

	for name, group := range sm.groups {
		stats.Groups[name] = GroupStats{
			SubscriberCount:  group.SubscriberCount(),
			BroadcasterCount: group.BroadcasterCount(),
		}
		stats.TotalSubscribers += group.SubscriberCount()
		stats.TotalBroadcasters += group.BroadcasterCount()
	}

	return stats
}

// Stats contains subscription statistics
type Stats struct {
	GroupCount         int                    // Number of groups
	TotalSubscribers   int                    // Total subscribers across all groups
	TotalBroadcasters  int                    // Total broadcasters across all groups
	Groups             map[string]GroupStats  // Per-group stats
}

// GroupStats contains statistics for a single group
type GroupStats struct {
	SubscriberCount  int // Number of subscribers
	BroadcasterCount int // Number of broadcasters
}
