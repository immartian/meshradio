# MeshRadio Multicast Overlay Specification

**Version:** 1.0
**Status:** Layer 4 (Multicast Overlay) - Core Implementation ✅
**Date:** 2025-12-03

## Overview

MeshRadio implements an **application-layer multicast overlay** that emulates multicast semantics using unicast transport over Yggdrasil. This enables both **regular multicast** (any-source) and **SSM** (source-specific multicast) for efficient one-to-many audio streaming.

## Motivation

**Problem:** Yggdrasil does NOT support IPv6 multicast routing.

**Solution:** Emulate multicast at the application layer:
- Subscribers explicitly subscribe to multicast groups
- Broadcasters maintain subscriber registries
- RTP packets are fan-out via unicast to each subscriber

## Architecture

### Regular Multicast (Any-Source)

**Use Case:** Emergency channels where listeners want to hear ALL broadcasters.

```
Group: "emergency"
├── Broadcaster A (201::1:8790)
├── Broadcaster B (201::2:8790)
└── Broadcaster C (201::3:8790)

Listener subscribes to group "emergency"
→ Receives RTP from A, B, and C
```

**Characteristics:**
- Listener subscribes to **group name** only
- Receives from **all broadcasters** in that group
- Similar to IPv6 regular multicast (ff02::)

### SSM (Source-Specific Multicast)

**Use Case:** Regular channels where listeners tune to a specific station.

```
Group: "community"
├── Broadcaster A (201::1:8795)
└── Broadcaster B (201::2:8795)

Listener subscribes to (201::1, "community")
→ Only receives from Broadcaster A
→ Ignores Broadcaster B
```

**Characteristics:**
- Listener subscribes to **(source, group)** pair
- Only receives from **specified broadcaster**
- Similar to IPv6 SSM (ff3x::)

## Protocol

### Subscription Packet

**Packet Type:** `PacketTypeSubscribe` (0x10)

**Payload:**
```
SubscribePayload {
    ListenerIPv6:  [16]byte  // Listener's IPv6 address
    ListenerPort:  uint16    // RTP port
    Callsign:      [16]byte  // Station callsign
    Group:         [32]byte  // Group name (e.g., "emergency")
    SSMSource:     [16]byte  // SSM source IPv6 (all zeros = regular multicast)
}
```

**Total Size:** 82 bytes

**Examples:**

Regular multicast subscription:
```go
SubscribePayload{
    ListenerIPv6: parseIPv6("201::100"),
    ListenerPort: 9001,
    Callsign:     "LISTENER-1",
    Group:        "emergency",
    SSMSource:    [16]byte{0}, // All zeros = regular multicast
}
```

SSM subscription:
```go
SubscribePayload{
    ListenerIPv6: parseIPv6("201::100"),
    ListenerPort: 9001,
    Callsign:     "LISTENER-1",
    Group:        "community",
    SSMSource:    parseIPv6("201::1"), // Only from this source
}
```

### Heartbeat Packet

**Packet Type:** `PacketTypeHeartbeat` (0x11)

**Purpose:** Keepalive to prevent subscription timeout

**Frequency:** Every 5 seconds

**Timeout:** 15 seconds (3 missed heartbeats)

### Unsubscribe Packet

**Packet Type:** `PacketTypeUnsubscribe` (0x12)

**Purpose:** Explicit unsubscribe (optional, timeout also works)

## Group Management

### Standard Groups

| Group Name  | Port Range | Purpose              | Multicast Type |
|-------------|------------|----------------------|----------------|
| emergency   | 8790-8794  | Emergency comms      | Regular        |
| community   | 8795-8797  | Community/public     | Regular/SSM    |
| talk        | 8798-8799  | General conversation | SSM            |

**Convention:**
- Emergency channels: **Regular multicast** (hear all broadcasters)
- Community/talk channels: **SSM** (tune to specific broadcaster)

### Creating Groups

Groups are created automatically when:
- First broadcaster registers
- First subscriber subscribes

Groups are deleted automatically when:
- Last subscriber unsubscribes
- Last broadcaster unregisters
- All members gone

## Subscription Flow

### Broadcaster Side

1. **Start broadcasting:**
   ```go
   sm := multicast.NewSubscriptionManager()

   broadcaster := &multicast.Broadcaster{
       IPv6:     myIPv6,
       Port:     8790,
       Callsign: "STATION-A",
   }

   sm.RegisterBroadcaster("emergency", broadcaster)
   ```

2. **Handle SUBSCRIBE packets:**
   ```go
   func handleSubscribe(packet *protocol.Packet) {
       payload := protocol.UnmarshalSubscribe(packet.Payload)

       subscriber := &multicast.Subscriber{
           IPv6:      payload.ListenerIPv6,
           Port:      payload.ListenerPort,
           Callsign:  payload.Callsign,
           SSMSource: payload.SSMSource,
       }

       sm.Subscribe(multicast.SubscribeRequest{
           Group:      protocol.GetGroupString(payload.Group),
           Subscriber: subscriber,
       })
   }
   ```

3. **Fan-out RTP packets:**
   ```go
   func streamAudio() {
       // Get subscribers for this broadcaster
       subs := sm.GetSubscribersForSource("emergency", myIPv6)

       for _, sub := range subs {
           addr := &net.UDPAddr{
               IP:   sub.IPv6,
               Port: sub.Port,
           }
           rtpSender.SendOpus(opusData, addr)
       }
   }
   ```

4. **Prune stale subscribers:**
   ```go
   ticker := time.NewTicker(5 * time.Second)
   for range ticker.C {
       sm.PruneStale(15 * time.Second)
   }
   ```

### Listener Side

1. **Subscribe to group (regular multicast):**
   ```go
   payload := protocol.SubscribePayload{
       ListenerIPv6: myIPv6,
       ListenerPort: 9001,
       Callsign:     "LISTENER-1",
       Group:        protocol.StringToGroup("emergency"),
       SSMSource:    [16]byte{0}, // Regular multicast
   }

   sendSubscribePacket(broadcasterAddr, payload)
   ```

2. **Subscribe to SSM:**
   ```go
   payload := protocol.SubscribePayload{
       ListenerIPv6: myIPv6,
       ListenerPort: 9001,
       Callsign:     "LISTENER-1",
       Group:        protocol.StringToGroup("community"),
       SSMSource:    protocol.IPv6ToBytes(broadcasterIPv6),
   }

   sendSubscribePacket(broadcasterAddr, payload)
   ```

3. **Send heartbeats:**
   ```go
   ticker := time.NewTicker(5 * time.Second)
   for range ticker.C {
       sendHeartbeat(broadcasterAddr)
   }
   ```

4. **Receive RTP:**
   ```go
   rtpReceiver := rtp.NewReceiver(9001)
   rtpReceiver.Start()

   for {
       opusData := rtpReceiver.ReadOpus()
       playAudio(opusData)
   }
   ```

## Discovery Integration

### With mDNS

Combine Layer 3 (mDNS Discovery) with Layer 4 (Multicast Overlay):

```go
// Discover emergency broadcasters
browser := mdns.NewBrowser()
services := browser.Browse(mdns.BrowseOptions{
    FilterChannel: "emergency",
})

// Subscribe to each (regular multicast)
for _, svc := range services {
    broadcasterAddr := &net.UDPAddr{
        IP:   svc.IPv6,
        Port: svc.Port,
    }

    sendSubscribe(broadcasterAddr, "emergency", nil)
}
```

### SSM with Specific Source

```go
// Find specific broadcaster
services := browser.Browse(mdns.BrowseOptions{
    FilterChannel: "community",
})

var targetBroadcaster mdns.ServiceInfo
for _, svc := range services {
    if svc.Callsign == "COMMUNITY-A" {
        targetBroadcaster = svc
        break
    }
}

// Subscribe to specific source (SSM)
sendSubscribe(targetBroadcaster.GetIPv6Addr(), "community", targetBroadcaster.IPv6)
```

## Fan-out Optimization

### Current Implementation (Sequential)

```go
for _, sub := range subscribers {
    rtpSender.SendOpus(data, sub.Addr)
}
```

**Pros:**
- Simple
- Preserves packet order
- Low overhead for small subscriber counts (<10)

**Cons:**
- Slower for large subscriber counts

### Future Optimization (Parallel)

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

**Pros:**
- Faster for large subscriber counts (>10)

**Cons:**
- More overhead (goroutines)
- Slightly more complex

**Recommendation:** Use sequential for now, optimize if needed.

## Performance Considerations

### Bandwidth

**Uplink bandwidth per broadcaster:**
```
Bandwidth = Audio Bitrate × Number of Subscribers
Example: 64 kbps × 10 subscribers = 640 kbps
```

**Scalability:**
- 1 broadcaster → 10 listeners: 640 kbps uplink
- 1 broadcaster → 100 listeners: 6.4 Mbps uplink
- 1 broadcaster → 1000 listeners: 64 Mbps uplink

**For large scale:** Use relay nodes (future enhancement)

### Latency

**Components:**
1. Audio capture: ~20ms
2. Opus encoding: ~5ms
3. RTP packetization: <1ms
4. Network transmission: variable (Yggdrasil routing)
5. Jitter buffer: ~20-50ms
6. Opus decoding: ~5ms
7. Audio playback: ~20ms

**Target:** <200ms end-to-end

## Testing

### Test Program

```bash
# Run multicast overlay test
./multicast-test
```

**Tests:**
- Regular multicast (all sources → listener)
- SSM (specific source → listener)
- Multiple listeners
- Statistics tracking
- Heartbeat/pruning

### Unit Tests

```bash
go test ./pkg/multicast/...
```

## Limitations

### Current Limitations

1. **No relay nodes:** Broadcaster must fan-out to all subscribers
2. **No rate limiting:** Subscription storms possible
3. **No QoS:** All subscribers treated equally
4. **No multicast routing:** Application-layer only

### Future Enhancements

- [ ] Relay nodes for scalability
- [ ] Rate limiting for SUBSCRIBE packets
- [ ] QoS tiers (priority subscribers)
- [ ] Mesh-wide multicast routing
- [ ] Anycast for load balancing

## Comparison with IPv6 Multicast

| Feature              | IPv6 Multicast     | MeshRadio Overlay |
|----------------------|--------------------|--------------------|
| Transport            | Network layer      | Application layer  |
| Group join           | MLD (IGMP)         | SUBSCRIBE packet   |
| Routing              | PIM, OSPF          | Unicast fan-out    |
| SSM support          | ff3x:: addresses   | SSMSource field    |
| Yggdrasil compatible | ❌ No              | ✅ Yes             |
| Overhead             | Low                | Medium             |
| Scalability          | Excellent          | Good               |

## Implementation Status

**Layer 4 (Multicast Overlay):**
- ✅ Subscription manager
- ✅ Group management
- ✅ Regular multicast support
- ✅ SSM support
- ✅ Heartbeat/pruning
- ✅ Statistics
- ✅ Protocol updates
- ✅ Test program
- ⏳ Broadcaster integration (future)
- ⏳ Listener integration (future)

---

**Next:** Integrate with Broadcaster/Listener (internal/)
