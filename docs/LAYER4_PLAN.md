# Layer 4: Multicast Overlay Implementation Plan

**Date:** 2025-12-03
**Goal:** Implement application-layer multicast overlay using unicast transport
**Duration:** ~3-4 hours

---

## Overview

Layer 4 builds the multicast overlay that emulates multicast semantics using Yggdrasil's unicast transport. This enables:

1. **Regular Multicast (Emergency)**: Any-source multicast for emergency channels
2. **SSM (Source-Specific Multicast)**: Source-specific for regular channels
3. **Subscription Management**: Track subscribers and manage group membership
4. **RTP Fan-out**: Distribute RTP packets to all subscribers via unicast

---

## Architecture

### Regular Multicast (Emergency Channels)

```
Emergency Channel "emergency"
├── Broadcaster A (203::1:8790)
├── Broadcaster B (203::2:8790)
└── Broadcaster C (203::3:8790)

Listener subscribes to group "emergency"
→ Discovers all broadcasters via mDNS
→ Subscribes to each broadcaster
→ Receives RTP from ALL sources
```

**Use Case:** Multiple stations broadcasting on emergency channel, listeners hear all.

### SSM (Regular Channels)

```
Community Channel "community"
├── Broadcaster A (203::1:8795)
└── Broadcaster B (203::2:8795)

Listener subscribes to (203::1, "community")
→ Only receives from Broadcaster A
→ Ignores Broadcaster B
```

**Use Case:** Listen to specific station, ignore others on same channel.

---

## Implementation Steps

### Part 1: Subscription Manager (45 min)

**Task 1.1: Create subscription package**

```
pkg/multicast/
  subscription.go   - Subscription management
  group.go          - Group management
  types.go          - Common types
```

**Task 1.2: Implement SubscriptionManager**

```go
type SubscriptionManager struct {
    groups map[string]*Group  // Key: group name (e.g., "emergency")
    mu     sync.RWMutex
}

type Group struct {
    Name        string
    Subscribers map[string]*Subscriber  // Key: IPv6:port
    Broadcasters map[string]*Broadcaster // Key: IPv6 (for regular multicast)
}

type Subscriber struct {
    IPv6      net.IP
    Port      int
    Callsign  string
    LastSeen  time.Time
    SSMSource net.IP  // nil = regular multicast, non-nil = SSM
}
```

**Task 1.3: Core methods**

```go
func (sm *SubscriptionManager) Subscribe(req SubscribeRequest) error
func (sm *SubscriptionManager) Unsubscribe(req UnsubscribeRequest) error
func (sm *SubscriptionManager) Heartbeat(ipv6 net.IP, port int) error
func (sm *SubscriptionManager) GetSubscribers(group string) []Subscriber
func (sm *SubscriptionManager) PruneStale(timeout time.Duration)
```

### Part 2: Group Management (45 min)

**Task 2.1: Group operations**

```go
func (sm *SubscriptionManager) CreateGroup(name string) error
func (sm *SubscriptionManager) DeleteGroup(name string) error
func (sm *SubscriptionManager) ListGroups() []string
func (sm *SubscriptionManager) GetGroupInfo(name string) (*Group, error)
```

**Task 2.2: Regular Multicast support**

```go
type SubscribeRequest struct {
    Group      string   // "emergency", "community", etc.
    Subscriber Subscriber
    SSMSource  net.IP   // nil = regular multicast
}

// Regular multicast: subscribe to group
sm.Subscribe(SubscribeRequest{
    Group: "emergency",
    Subscriber: sub,
    SSMSource: nil,  // Receive from ALL sources
})

// SSM: subscribe to (source, group)
sm.Subscribe(SubscribeRequest{
    Group: "community",
    Subscriber: sub,
    SSMSource: sourceIPv6,  // Only receive from this source
})
```

**Task 2.3: Broadcaster registration**

```go
func (sm *SubscriptionManager) RegisterBroadcaster(group string, ipv6 net.IP) error
func (sm *SubscriptionManager) UnregisterBroadcaster(group string, ipv6 net.IP) error
func (sm *SubscriptionManager) GetBroadcasters(group string) []net.IP
```

### Part 3: Integration with Broadcaster (45 min)

**Task 3.1: Update internal/broadcaster/broadcaster.go**

```go
type Broadcaster struct {
    // Existing fields...
    subManager *multicast.SubscriptionManager
    rtpSender  *rtp.Sender
}

func (b *Broadcaster) Start() error {
    // Start RTP sender
    // Register with subscription manager
    // Start subscription handler
    // Start RTP streaming loop
}

func (b *Broadcaster) handleSubscription(packet *protocol.Packet) error {
    req := parseSubscribePacket(packet)

    // Add to subscription manager
    b.subManager.Subscribe(req)

    // Send confirmation
    sendAck(req.Subscriber)
}

func (b *Broadcaster) streamingLoop() {
    for {
        // Get audio frame
        opusData := b.getAudioFrame()

        // Get subscribers for this group
        subs := b.subManager.GetSubscribers(b.group)

        // Send RTP to each subscriber
        for _, sub := range subs {
            addr := &net.UDPAddr{
                IP:   sub.IPv6,
                Port: sub.Port,
            }
            b.rtpSender.SendOpus(opusData, addr)
        }
    }
}
```

**Task 3.2: Update internal/listener/listener.go**

```go
type Listener struct {
    // Existing fields...
    rtpReceiver *rtp.Receiver
    ssmSource   net.IP  // nil = regular multicast
}

func (l *Listener) Subscribe(group string, ssmSource net.IP) error {
    l.ssmSource = ssmSource

    // Send SUBSCRIBE packet
    req := protocol.SubscribePayload{
        ListenerIPv6: l.localIPv6,
        ListenerPort: l.localPort,
        Callsign:     l.callsign,
        Group:        group,
        SSMSource:    ssmSource,
    }

    l.sendSubscribe(req)

    // Start RTP receiver
    l.rtpReceiver.Start()

    // Start heartbeat
    l.startHeartbeat()
}
```

### Part 4: Protocol Updates (30 min)

**Task 4.1: Update pkg/protocol/subscription.go**

Add SSM support to subscription payload:

```go
type SubscribePayload struct {
    ListenerIPv6 [16]byte
    ListenerPort uint16
    Callsign     [16]byte
    Group        [32]byte  // NEW: Group name
    SSMSource    [16]byte  // NEW: SSM source (all zeros = regular multicast)
}
```

**Task 4.2: Update packet types**

Already have:
- PacketTypeSubscribe (0x10)
- PacketTypeHeartbeat (0x11)

Add:
- PacketTypeUnsubscribe (0x12)

### Part 5: Testing (45 min)

**Task 5.1: Create test program**

```
cmd/multicast-test/
  main.go - Test multicast overlay
```

Modes:
- `broadcast` - Start broadcaster on group
- `listen-all` - Regular multicast (listen to all sources)
- `listen-ssm` - SSM (listen to specific source)

**Task 5.2: Test scenarios**

**Scenario 1: Regular Multicast (Emergency)**
```bash
# Terminal 1: Broadcaster A
./multicast-test -mode broadcast -group emergency -callsign STATION-A -port 8790

# Terminal 2: Broadcaster B
./multicast-test -mode broadcast -group emergency -callsign STATION-B -port 8790

# Terminal 3: Listener (receives from BOTH)
./multicast-test -mode listen-all -group emergency
```

**Scenario 2: SSM (Community)**
```bash
# Terminal 1: Broadcaster A
./multicast-test -mode broadcast -group community -callsign STATION-A -port 8795

# Terminal 2: Broadcaster B
./multicast-test -mode broadcast -group community -callsign STATION-B -port 8795

# Terminal 3: Listener (receives only from A)
./multicast-test -mode listen-ssm -group community -source 203:abcd::1
```

**Scenario 3: Multiple Listeners**
```bash
# Terminal 1: Broadcaster
./multicast-test -mode broadcast -group community -callsign STATION-A -port 8795

# Terminal 2-4: Listeners
./multicast-test -mode listen-ssm -group community -source 203:abcd::1
```

---

## Code Structure

```
pkg/multicast/
  subscription.go   - SubscriptionManager
  group.go          - Group management
  types.go          - Common types (Subscriber, Group, etc.)

internal/broadcaster/
  broadcaster.go    - Updated with multicast support

internal/listener/
  listener.go       - Updated with SSM support

pkg/protocol/
  subscription.go   - Updated with Group and SSMSource fields
  packet.go         - Add PacketTypeUnsubscribe

cmd/multicast-test/
  main.go           - Test program
```

---

## Key Design Decisions

### 1. Regular Multicast vs SSM

**Regular Multicast (Emergency):**
- Listener subscribes to **group only**
- Receives from **all broadcasters** in that group
- Use case: Emergency channels where you want to hear everyone

**SSM (Regular Channels):**
- Listener subscribes to **(source, group)** pair
- Receives only from **specific broadcaster**
- Use case: Regular channels where you tune to one station

### 2. Discovery Integration

**With mDNS:**
```go
// Discover all emergency broadcasters
browser := mdns.NewBrowser()
services := browser.Browse(mdns.BrowseOptions{
    FilterChannel: "emergency",
})

// Subscribe to all (regular multicast)
for _, svc := range services {
    listener.Subscribe("emergency", nil)  // nil = all sources
}
```

**SSM with known source:**
```go
// Subscribe to specific source
listener.Subscribe("community", sourceIPv6)
```

### 3. Fan-out Optimization

**Current: Sequential send**
```go
for _, sub := range subscribers {
    rtpSender.SendOpus(data, sub.Addr)
}
```

**Future: Parallel send (if needed)**
```go
var wg sync.WaitGroup
for _, sub := range subscribers {
    wg.Add(1)
    go func(addr *net.UDPAddr) {
        defer wg.Done()
        rtpSender.SendOpus(data, addr)
    }(sub.Addr)
}
wg.Wait()
```

### 4. Stale Subscriber Cleanup

```go
// Prune subscribers with no heartbeat for 15 seconds
ticker := time.NewTicker(5 * time.Second)
for range ticker.C {
    subManager.PruneStale(15 * time.Second)
}
```

---

## Success Criteria

By end of this session:

- [ ] `pkg/multicast/` package implemented
- [ ] Broadcaster supports group-based fan-out
- [ ] Listener supports SSM and regular multicast
- [ ] Protocol updated with Group and SSMSource fields
- [ ] Test program demonstrates both modes
- [ ] Regular multicast works (multiple sources → one listener)
- [ ] SSM works (specific source → listener)
- [ ] Multiple listeners receive streams

---

## Testing Checklist

### Unit Tests
- [ ] SubscriptionManager.Subscribe()
- [ ] SubscriptionManager.Unsubscribe()
- [ ] SubscriptionManager.PruneStale()
- [ ] Group.AddSubscriber()
- [ ] Group.RemoveSubscriber()

### Integration Tests
- [ ] Broadcaster advertises via mDNS
- [ ] Listener discovers and subscribes
- [ ] RTP packets reach subscriber
- [ ] Heartbeat keeps subscription alive
- [ ] Timeout removes stale subscriber

### System Tests
- [ ] 2 broadcasters + 1 listener (regular multicast)
- [ ] 1 broadcaster + 3 listeners (SSM)
- [ ] Mixed: some regular, some SSM on same group

---

## Potential Issues

### Issue 1: RTP timestamp sync (multiple sources)
**Problem:** Different broadcasters have different timestamps
**Solution:** Receiver handles per-source jitter buffer
**Defer:** Phase 2 - for now, accept out-of-order from multiple sources

### Issue 2: Group name conflicts
**Problem:** Different meanings for same group name
**Solution:** Use channel naming convention (emergency/community/talk)
**Mitigation:** Document standard group names in spec

### Issue 3: Subscription storms (many listeners)
**Problem:** 100+ listeners subscribe at once
**Solution:** Rate limit SUBSCRIBE packets
**Defer:** Optimize only if needed

### Issue 4: NAT traversal for RTP
**Problem:** RTP port may not be reachable
**Solution:** Yggdrasil handles NAT traversal for us
**Note:** This just works™ on Yggdrasil

---

## Documentation Updates

### Update TODO.md
- Mark Layer 4 tasks as complete
- Update status to ✅

### Create MULTICAST_SPEC.md
Document:
- Regular multicast vs SSM
- Group naming convention
- Subscription protocol
- Example code

### Update NEXT_SESSION.md
Plan for Layer 5 (Emergency Features)

---

## After This Session

**What we'll have:**
- ✅ Application-layer multicast overlay
- ✅ Regular multicast for emergency channels
- ✅ SSM for regular channels
- ✅ Subscription management with heartbeat
- ✅ RTP fan-out to multiple listeners

**What's next:**
- Layer 5: Emergency features (priority, auto-tune, net control)
- Layer 6: Service layer API
- Layer 7: UI integration

---

## Quick Start

```bash
# 1. Create package structure
mkdir -p pkg/multicast

# 2. Implement core types
touch pkg/multicast/{types,subscription,group}.go

# 3. Update protocol
# Edit pkg/protocol/subscription.go

# 4. Update broadcaster/listener
# Edit internal/{broadcaster,listener}/*.go

# 5. Create test program
mkdir -p cmd/multicast-test
touch cmd/multicast-test/main.go

# 6. Build and test
go build -o multicast-test ./cmd/multicast-test
./multicast-test -mode broadcast -group emergency
```

---

**Ready to build Layer 4!** 🚀
